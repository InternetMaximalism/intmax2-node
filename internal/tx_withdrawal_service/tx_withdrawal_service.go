package tx_transfer_service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTree "intmax2-node/internal/tree"
	"intmax2-node/internal/tx_transfer_service"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func WithdrawalTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	args []string,
	amountStr string,
	recipientAddressHex string,
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

	log.Infof("User's INTMAX Address: %s\n", userAccount.ToAddress().String())
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
	recipientBytes, err := hexutil.Decode(recipientAddressHex)
	if err != nil {
		log.Fatalf("failed to parse recipient address: %v", err)
	}

	recipientAddress, err := intMaxTypes.NewEthereumAddress(recipientBytes)
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
	const transferIndex = 0
	initialLeaves[transferIndex] = transfer

	transferTree, err := intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, initialLeaves, zeroTransfer.Hash())
	if err != nil {
		log.Fatalf("failed to create transfer tree: %v", err)
	}

	transferMerkleProof, transfersHash, err := transferTree.ComputeMerkleProof(0)
	if err != nil {
		log.Fatalf("failed to compute merkle proof: %v", err)
	}

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

	emptyBackupTx := transaction.BackupTransactionData{
		EncodedEncryptedTx: "",
		Signature:          "0x",
	}
	emptyBackupTransfers := make([]*transaction.BackupTransferInput, len(initialLeaves))
	err = SendWithdrawalTransaction(
		ctx,
		cfg,
		log,
		userAccount,
		transfersHash,
		nonce,
		&emptyBackupTx,
		emptyBackupTransfers,
	)
	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	}

	log.Printf("The transaction request has been successfully sent. Please wait for the server's response.")

	// Get proposed block
	proposedBlock, err := tx_transfer_service.GetBlockProposed(
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

	encodedEncryptedTx := base64.StdEncoding.EncodeToString(encryptedTx)
	backupTx := transaction.BackupTransactionData{
		EncodedEncryptedTx: encodedEncryptedTx,
		Signature:          "0x",
	}

	backupTransfers := make([]*transaction.BackupTransferInput, len(initialLeaves))
	for i, transfer := range initialLeaves {
		backupTransfers[i], err = tx_transfer_service.MakeWithdrawalBackupData(
			transfer, userAccount.ToAddress(), transfersHash, nonce, proposedBlock, transferMerkleProof,
		)
		if err != nil {
			log.Fatalf("failed to make backup data: %v", err)
		}
	}

	// Accept proposed block
	err = tx_transfer_service.SendSignedProposedBlock(
		ctx, cfg, log, userAccount, proposedBlock.TxTreeRoot, *txHash, proposedBlock.PublicKeys,
		&backupTx, backupTransfers,
	)
	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	}

	log.Infof("The transaction has been successfully sent.")

	// TODO: Make it safe to interrupt here.

	publicKeysStr := make([]string, len(proposedBlock.PublicKeys))
	for i, key := range proposedBlock.PublicKeys {
		publicKeysStr[i] = key.ToAddress().String()
	}

	txIndex := slices.Index(publicKeysStr, userAccount.ToAddress().String())
	if txIndex == -1 {
		log.Fatalf("failed to find user's public key in the proposed block")
	}

	err = SendWithdrawalTransactionFromBackupTransfer(
		ctx, cfg, log, backupTransfers[transferIndex],
	)

	// // Send withdrawal request
	// err = SendWithdrawalRequest(
	// 	ctx, cfg, log,
	// 	userAccount.ToAddress(),
	// 	transfer,
	// 	transferMerkleProof,
	// 	transferIndex,
	// 	transfersHash,
	// 	nonce,
	// 	proposedBlock.TxTreeMerkleProof,
	// 	int32(txIndex),
	// 	proposedBlock.TxTreeRoot,
	// )
	if err != nil {
		log.Fatalf("failed to request withdrawal: %v", err)
	}

	log.Infof("The withdrawal request has been successfully sent.")
}

func SendWithdrawalTransactionFromBackupTransfer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	backupWithdrawal *transaction.BackupTransferInput,
) error {
	// base64 decode
	encodedEncryptedTransfer, err := base64.StdEncoding.DecodeString(backupWithdrawal.EncodedEncryptedTransfer)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// json unmarshal
	var withdrawal tx_transfer_service.Withdrawal
	err = json.Unmarshal(encodedEncryptedTransfer, &withdrawal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	proposedBlock := tx_transfer_service.BlockProposedResponseData{
		TxTreeRoot:        withdrawal.TxTreeRoot,
		TxTreeMerkleProof: withdrawal.TxTreeMerkleProof,
	}

	recipientBytes := withdrawal.Transfer.Recipient.Bytes()
	recipient, err := intMaxTypes.NewEthereumAddress(recipientBytes)
	if err != nil {
		return fmt.Errorf("failed to create recipient address: %w", err)
	}
	amount, ok := new(big.Int).SetString(withdrawal.Transfer.Amount, 10)
	if !ok {
		return fmt.Errorf("failed to convert amount to int: %v", withdrawal.Transfer.Amount)
	}
	transfer := intMaxTypes.NewTransfer(
		recipient,
		withdrawal.Transfer.TokenIndex,
		amount,
		withdrawal.Transfer.Salt,
	)

	nonce, err := strconv.ParseUint(withdrawal.Nonce, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse nonce: %w", err)
	}

	senderAddress, err := intMaxAcc.NewAddressFromHex(withdrawal.SenderAddress)
	if err != nil {
		return fmt.Errorf("failed to parse sender address: %w", err)
	}

	// Send withdrawal transaction
	return SendWithdrawalRequest(
		ctx, cfg, log,
		senderAddress,
		transfer,
		withdrawal.TransferMerkleProof,
		withdrawal.TransferIndex,
		withdrawal.TransferTreeRoot,
		nonce,
		proposedBlock.TxTreeMerkleProof,
		withdrawal.TxIndex,
		proposedBlock.TxTreeRoot,
	)
}

func SendWithdrawalRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	senderAddress intMaxAcc.Address,
	transfer *intMaxTypes.Transfer,
	transferMerkleProof []*intMaxTypes.PoseidonHashOut,
	transferIndex int32,
	TransferTreeRoot intMaxTypes.PoseidonHashOut,
	nonce uint64,
	txTreeMerkleProof []*goldenposeidon.PoseidonHashOut,
	txIndex int32,
	txTreeRoot goldenposeidon.PoseidonHashOut,
) error {
	// TODO: Get the block number and block hash
	// Specify the block number containing the transaction.
	rollupCfg := intMaxTypes.NewRollupContractConfigFromEnv(cfg, "https://sepolia-rpc.scroll.io")
	blockNumber, err := intMaxTypes.FetchLatestIntMaxBlockNumber(rollupCfg, ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch latest intmax block: %w", err)
	}
	blockHash, err := intMaxTypes.FetchBlockHash(rollupCfg, ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch block hash: %w", err)
	}

	err = SendWithdrawalWithRawRequest(
		ctx, cfg, log, transfer, TransferTreeRoot, nonce, transferMerkleProof, transferIndex, txTreeMerkleProof, int32(txIndex),
		blockNumber, blockHash,
	)
	if err != nil {
		return fmt.Errorf("failed to request withdrawal: %w", err)
	}

	return nil
}
