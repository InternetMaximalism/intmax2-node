package tx_transfer_service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxAccTypes "intmax2-node/internal/accounts/types"
	"intmax2-node/internal/balance_service"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	MyTransferIndex = 0 // TODO: 1
	base10Key       = 10
	uint64Key       = 64
)

func TransferTransaction(
	ctx context.Context,
	cfg *configs.Config,
	// log logger.Logger,
	sb ServiceBlockchain,
	args []string,
	amountStr string,
	recipientAddressStr string,
	userEthPrivateKey string,
) error {
	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		return fmt.Errorf("fail to parse user private key: %v", err)
	}

	userAccount, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		return fmt.Errorf("fail to parse user private key: %v", err)
	}

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	tokenIndex, err := balance_service.GetTokenIndexFromLiquidityContract(ctx, cfg, sb, *tokenInfo)
	if err != nil {
		return err
	}

	fmt.Printf("User's INTMAX Address: %s\n", userAccount.ToAddress().String())
	fmt.Println("Fetching balances...")
	balance, err := balance_service.GetUserBalance(ctx, cfg, userAccount, tokenIndex)
	if err != nil {
		return fmt.Errorf(ErrFailedToGetBalance.Error()+": %v", err)
	}

	if strings.TrimSpace(amountStr) == "" {
		return fmt.Errorf("amount is required")
	}

	const int10Key = 10
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %v", amountStr)
	}

	if balance.Cmp(amount) < 0 {
		return fmt.Errorf("insufficient balance: %s", balance)
	}

	// Send transfer transaction
	recipient, err := intMaxAcc.NewPublicKeyFromAddressHex(recipientAddressStr)
	if err != nil {
		return fmt.Errorf("failed to parse recipient address: %v", err)
	}

	recipientAddress, err := intMaxTypes.NewINTMAXAddress(recipient.ToAddress().Bytes())
	if err != nil {
		return fmt.Errorf("failed to create recipient address: %v", err)
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
		return fmt.Errorf("failed to create transfer tree: %v", err)
	}

	transfersHash, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	var nonce uint64 = 1 // TODO: Incremented with each transaction

	err = SendTransferTransaction(
		ctx,
		cfg,
		// log,
		userAccount,
		transfersHash,
		nonce,
		// &backupTx,
		// backupTransfers,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Println("The transaction request has been successfully sent. Please wait for the server's response.")

	// Get proposed block
	proposedBlock, err := GetBlockProposed(
		ctx, cfg, userAccount, transfersHash, nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Println("The proposed block has been successfully received.")

	tx, err := intMaxTypes.NewTx(
		&transfersHash,
		nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to create new tx: %w", err)
	}

	txHash := tx.Hash()

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
		return fmt.Errorf("failed to encrypt deposit: %w", err)
	}

	encodedEncryptedTx := base64.StdEncoding.EncodeToString(encryptedTx)
	backupTx := transaction.BackupTransactionData{
		EncodedEncryptedTx: encodedEncryptedTx,
		Signature:          "0x",
	}

	backupTransfers := make([]*transaction.BackupTransferInput, len(initialLeaves))
	for i := range initialLeaves {
		backupTransfers[i], err = MakeTransferBackupData(initialLeaves[i])
		if err != nil {
			return fmt.Errorf("failed to make backup data: %v", err)
		}
	}

	// Accept proposed block
	err = SendSignedProposedBlock(
		ctx, cfg, userAccount, proposedBlock.TxTreeRoot, *txHash, proposedBlock.PublicKeys,
		&backupTx, backupTransfers,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Println("The transaction has been successfully sent.")

	return nil
}

var ErrFailedToCreateRecipientAddress = errors.New("failed to create recipient address")
var ErrFailedToGetRecipientPublicKey = errors.New("failed to get recipient public key")
var ErrFailedToEncryptTransfer = errors.New("failed to encrypt transfer")

func MakeTransferBackupData(transfer *intMaxTypes.Transfer) (backupTransfer *transaction.BackupTransferInput, _ error) {
	if transfer.Recipient.TypeOfAddress != "INTMAX" {
		return nil, errors.New("recipient address should be INTMAX")
	}

	recipientAddress, err := transfer.Recipient.ToINTMAXAddress()
	if err != nil {
		return nil, errors.Join(ErrFailedToCreateRecipientAddress, err)
	}
	recipientPublicKey, err := recipientAddress.Public()
	if err != nil {
		return nil, errors.Join(ErrFailedToGetRecipientPublicKey, err)
	}

	encryptedTransfer, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		recipientPublicKey,
		transfer.Marshal(),
	)
	if err != nil {
		return nil, errors.Join(ErrFailedToEncryptTransfer, err)
	}

	return &transaction.BackupTransferInput{
		Recipient:                hexutil.Encode(transfer.Recipient.Marshal()),
		EncodedEncryptedTransfer: base64.StdEncoding.EncodeToString(encryptedTransfer),
	}, nil
}

type BackupWithdrawal struct {
	SenderAddress       intMaxAcc.Address                 `json:"senderAddress"`
	Transfer            *intMaxTypes.Transfer             `json:"transfer"`
	TransferMerkleProof []*intMaxTypes.PoseidonHashOut    `json:"transferMerkleProof"`
	TransferIndex       int32                             `json:"transferIndex"`
	TransferTreeRoot    intMaxTypes.PoseidonHashOut       `json:"transferTreeRoot"`
	Nonce               uint64                            `json:"nonce"`
	TxTreeMerkleProof   []*goldenposeidon.PoseidonHashOut `json:"txTreeMerkleProof"`
	TxIndex             int32                             `json:"txIndex"`
	TxTreeRoot          goldenposeidon.PoseidonHashOut    `json:"txTreeRoot"`
}

type withdrawalTransfer struct {
	Recipient common.Address `json:"recipient"`

	TokenIndex uint32 `json:"tokenIndex"`

	// Amount is a decimal string
	Amount string `json:"amount"`

	Salt *goldenposeidon.PoseidonHashOut `json:"salt"`
}

type backupWithdrawal struct {
	SenderAddress string `json:"senderAddress"`

	Transfer withdrawalTransfer `json:"transfer"`

	TransferMerkleProof []*goldenposeidon.PoseidonHashOut `json:"transferMerkleProof"`

	TransferIndex int32 `json:"transferIndex"`

	TransferTreeRoot goldenposeidon.PoseidonHashOut `json:"transferTreeRoot"`

	// Nonce is a decimal string
	Nonce string `json:"nonce"`

	TxTreeMerkleProof []*goldenposeidon.PoseidonHashOut `json:"txTreeMerkleProof"`

	TxIndex int32 `json:"txIndex"`

	TxTreeRoot goldenposeidon.PoseidonHashOut `json:"txTreeRoot"`
}

func (bw *BackupWithdrawal) MarshalJSON() ([]byte, error) {
	if bw.Transfer.Recipient.TypeOfAddress != intMaxAccTypes.EthereumAddressType {
		return nil, errors.New("recipient address should be ETHEREUM")
	}

	recipient, err := bw.Transfer.Recipient.ToEthereumAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to convert recipient address: %w", err)
	}

	return json.Marshal(&backupWithdrawal{
		SenderAddress: bw.SenderAddress.String(),
		Transfer: withdrawalTransfer{
			Recipient:  recipient,
			TokenIndex: bw.Transfer.TokenIndex,
			Amount:     bw.Transfer.Amount.String(),
			Salt:       bw.Transfer.Salt,
		},
		TransferMerkleProof: bw.TransferMerkleProof,
		TransferIndex:       bw.TransferIndex,
		TransferTreeRoot:    bw.TransferTreeRoot,
		Nonce:               strconv.FormatUint(bw.Nonce, base10Key),
		TxTreeMerkleProof:   bw.TxTreeMerkleProof,
		TxIndex:             bw.TxIndex,
		TxTreeRoot:          bw.TxTreeRoot,
	})
}

func (bw *BackupWithdrawal) UnmarshalJSON(data []byte) error {
	var withdrawal backupWithdrawal
	err := json.Unmarshal(data, &withdrawal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	recipientBytes := withdrawal.Transfer.Recipient.Bytes()
	recipient, err := intMaxTypes.NewEthereumAddress(recipientBytes)
	if err != nil {
		return fmt.Errorf("failed to create recipient address: %w", err)
	}
	amount, ok := new(big.Int).SetString(withdrawal.Transfer.Amount, base10Key)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %v", withdrawal.Transfer.Amount)
	}
	transfer := intMaxTypes.NewTransfer(
		recipient,
		withdrawal.Transfer.TokenIndex,
		amount,
		withdrawal.Transfer.Salt,
	)

	nonce, err := strconv.ParseUint(withdrawal.Nonce, base10Key, uint64Key)
	if err != nil {
		return fmt.Errorf("failed to parse nonce: %w", err)
	}

	senderAddress, err := intMaxAcc.NewAddressFromHex(withdrawal.SenderAddress)
	if err != nil {
		return fmt.Errorf("failed to parse sender address: %w", err)
	}

	bw.SenderAddress = senderAddress
	bw.Transfer = transfer
	bw.TransferMerkleProof = withdrawal.TransferMerkleProof
	bw.TransferIndex = withdrawal.TransferIndex
	bw.TransferTreeRoot = withdrawal.TransferTreeRoot
	bw.Nonce = nonce
	bw.TxTreeMerkleProof = withdrawal.TxTreeMerkleProof
	bw.TxIndex = withdrawal.TxIndex
	bw.TxTreeRoot = withdrawal.TxTreeRoot

	return nil
}

func MakeWithdrawalBackupData(
	transfer *intMaxTypes.Transfer,
	senderAddress intMaxAcc.Address,
	transfersHash goldenposeidon.PoseidonHashOut,
	nonce uint64,
	txTreeRoot goldenposeidon.PoseidonHashOut,
	txTreeMerkleProof []*goldenposeidon.PoseidonHashOut,
	transferMerkleProof []*goldenposeidon.PoseidonHashOut,
	txIndex int32,
	transferIndex int32,
) (backupTransfer *transaction.BackupTransferInput, _ error) {
	if transfer.Recipient.TypeOfAddress != "ETHEREUM" {
		return nil, errors.New("recipient address should be ETHEREUM")
	}

	recipient, err := transfer.Recipient.ToEthereumAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to create recipient address: %w", err)
	}

	withdrawal := backupWithdrawal{
		SenderAddress: senderAddress.String(),
		Transfer: withdrawalTransfer{
			Recipient:  recipient,
			TokenIndex: transfer.TokenIndex,
			Amount:     transfer.Amount.String(),
			Salt:       transfer.Salt,
		},
		TransferMerkleProof: transferMerkleProof,
		TransferTreeRoot:    transfersHash,
		TransferIndex:       transferIndex,
		Nonce:               strconv.FormatUint(nonce, base10Key),
		TxTreeMerkleProof:   txTreeMerkleProof,
		TxTreeRoot:          txTreeRoot,
		TxIndex:             txIndex,
	}

	// No encryption
	encryptedTransfer, err := json.Marshal(&withdrawal)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	encryptedTransfer2, err := json.Marshal(&BackupWithdrawal{
		SenderAddress:       senderAddress,
		Transfer:            transfer,
		TransferMerkleProof: transferMerkleProof,
		TransferIndex:       transferIndex,
		TransferTreeRoot:    transfersHash,
		Nonce:               nonce,
		TxTreeMerkleProof:   txTreeMerkleProof,
		TxIndex:             txIndex,
		TxTreeRoot:          txTreeRoot,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if !bytes.Equal(encryptedTransfer, encryptedTransfer2) {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return &transaction.BackupTransferInput{
		Recipient:                hexutil.Encode(transfer.Recipient.Marshal()),
		EncodedEncryptedTransfer: base64.StdEncoding.EncodeToString(encryptedTransfer),
	}, nil
}
