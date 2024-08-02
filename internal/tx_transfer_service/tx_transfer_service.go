package tx_transfer_service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const MyTransferIndex = 0 // TODO: 1

func SendTransferTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	args []string,
	amountStr string,
	recipientAddressStr string,
	userEthPrivateKey string,
) {
	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		log.Fatalf("fail to parse user private key: %v", err)
	}

	userAccount, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		log.Fatalf("fail to parse user private key: %v", err)
	}

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		log.Fatalf("%s", err)
	}

	tokenIndex, err := balance_service.GetTokenIndexFromLiquidityContract(ctx, cfg, sb, *tokenInfo)
	if err != nil {
		log.Fatalf("%s", errors.Join(ErrTokenNotFound, err))
	}

	balance, err := balance_service.GetUserBalance(ctx, cfg, log, userAccount, tokenIndex)
	if err != nil {
		log.Fatalf(ErrFailedToGetBalance+": %v", err)
	}

	if strings.TrimSpace(amountStr) == "" {
		log.Fatalf("Amount is required")
	}

	const int10Key = 10
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		log.Fatalf("failed to convert amount to int: %v", amountStr)
	}

	if balance.Cmp(amount) < 0 {
		log.Fatalf("Insufficient balance: %s", balance)
	}

	// Send transfer transaction
	recipient, err := intMaxAcc.NewPublicKeyFromAddressHex(recipientAddressStr)
	if err != nil {
		log.Fatalf("failed to parse recipient address: %v", err)
	}

	recipientAddress, err := intMaxTypes.NewINTMAXAddress(recipient.ToAddress().Bytes())
	if err != nil {
		log.Fatalf("failed to create recipient address: %v", err)
	}

	transfer := intMaxTypes.NewTransferWithRandomSalt(
		recipientAddress,
		tokenIndex,
		amount,
	)

	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	initialLeaves := make([]*intMaxTypes.Transfer, 1)
	initialLeaves[MyTransferIndex] = transfer

	transferTree, err := intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, initialLeaves, zeroTransfer.Hash())
	if err != nil {
		log.Fatalf("failed to create transfer tree: %v", err)
	}

	transfersHash, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	var nonce uint64 = 1 // TODO: Incremented with each transaction

	txDetails := intMaxTypes.TxDetails{
		Tx: intMaxTypes.Tx{
			TransferTreeRoot: &transfersHash,
			Nonce:            nonce,
		},
		Transfers: initialLeaves,
	}

	encodedTx := txDetails.Marshal()
	encryptedTx, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		userAccount.Public(),
		encodedTx,
	)
	if err != nil {
		log.Errorf("failed to encrypt deposit: %w", err)
		return
	}

	encodedEncryptedTx := base64.StdEncoding.EncodeToString(encryptedTx)
	backupTransfers, err := MakeBackupData(initialLeaves)
	if err != nil {
		log.Fatalf("failed to make backup data: %v", err)
	}

	backupTx := transaction.BackupTransactionData{
		EncodedEncryptedTx: encodedEncryptedTx,
		Signature:          "0x",
	}

	err = SendTransactionRequest(
		ctx,
		cfg,
		log,
		userAccount,
		transfersHash,
		nonce,
		&backupTx,
		backupTransfers,
	)
	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	}

	log.Printf("The transaction request has been successfully sent. Please wait for the server's response.")

	// Get proposed block
	proposedBlock, err := GetBlockProposed(
		ctx, cfg, log, userAccount, transfersHash, nonce,
	)
	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	}

	log.Infof("The proposed block has been successfully received. Please wait for the server's response.")

	tx, err := intMaxTypes.NewTx(
		&transfersHash,
		nonce,
	)
	if err != nil {
		log.Fatalf("failed to create new tx: %w", err)
	}

	txHash := tx.Hash()

	// Accept proposed block
	err = SendSignedProposedBlock(
		ctx, cfg, log, userAccount, proposedBlock.TxTreeRoot, *txHash, proposedBlock.PublicKeys,
	)
	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	}

	log.Printf("The transaction has been successfully sent.")
}

func MakeBackupData(transfers []*intMaxTypes.Transfer) (encodedEncryptedTransfers []*transaction.BackupTransferInput, _ error) {
	var ErrFailedToCreateRecipientAddress = errors.New("failed to create recipient address")
	var ErrFailedToGetRecipientPublicKey = errors.New("failed to get recipient public key")
	var ErrFailedToEncryptTransfer = errors.New("failed to encrypt transfer")

	encodedEncryptedTransfers = make([]*transaction.BackupTransferInput, len(transfers))
	for i, transfer := range transfers {
		var encryptedTransfer []byte
		if transfer.Recipient.TypeOfAddress == "INTMAX" {
			recipientAddress, err := transfer.Recipient.ToINTMAXAddress()
			if err != nil {
				return nil, errors.Join(ErrFailedToCreateRecipientAddress, err)
			}
			recipientPublicKey, err := recipientAddress.Public()
			if err != nil {
				return nil, errors.Join(ErrFailedToGetRecipientPublicKey, err)
			}

			encryptedTransfer, err = intMaxAcc.EncryptECIES(
				rand.Reader,
				recipientPublicKey,
				transfer.Marshal(),
			)
			if err != nil {
				return nil, errors.Join(ErrFailedToEncryptTransfer, err)
			}
		} else {
			// No encryption
			encryptedTransfer = transfer.Marshal()
		}

		encodedEncryptedTransfers[i] = &transaction.BackupTransferInput{
			Recipient:                hexutil.Encode(transfer.Recipient.Marshal()),
			EncodedEncryptedTransfer: base64.StdEncoding.EncodeToString(encryptedTransfer),
		}
	}

	return encodedEncryptedTransfers, nil
}
