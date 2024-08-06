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
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTree "intmax2-node/internal/tree"
	"intmax2-node/internal/tx_transfer_service"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"math/big"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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

	err = SendWithdrawalTransaction(
		ctx,
		cfg,
		log,
		userAccount,
		transfersHash,
		nonce,
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

	log.Infof("The proposed block has been successfully received.")

	tx, err := intMaxTypes.NewTx(
		&transfersHash,
		nonce,
	)
	if err != nil {
		log.Fatalf("failed to create new tx: %w", err)
	}

	txHash := tx.Hash()

	publicKeysStr := make([]string, len(proposedBlock.PublicKeys))
	for i, key := range proposedBlock.PublicKeys {
		publicKeysStr[i] = key.ToAddress().String()
	}

	txIndex := slices.Index(publicKeysStr, userAccount.ToAddress().String())
	if txIndex == -1 {
		log.Fatalf("failed to find user's public key in the proposed block")
	}

	encodedEncryptedTx := base64.StdEncoding.EncodeToString(encryptedTx)
	backupTx := transaction.BackupTransactionData{
		EncodedEncryptedTx: encodedEncryptedTx,
		Signature:          "0x",
	}

	backupTransfers := make([]*transaction.BackupTransferInput, len(initialLeaves))
	for i, transfer := range initialLeaves {
		backupTransfers[i], err = tx_transfer_service.MakeWithdrawalBackupData(
			transfer,
			userAccount.ToAddress(),
			transfersHash,
			nonce,
			proposedBlock.TxTreeRoot,
			proposedBlock.TxTreeMerkleProof,
			transferMerkleProof,
			transferIndex,
			int32(txIndex),
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

	// Send withdrawal request
	err = SendWithdrawalRequest(ctx, cfg, log, &tx_transfer_service.BackupWithdrawal{
		SenderAddress:       userAccount.ToAddress(),
		Transfer:            transfer,
		TransferMerkleProof: transferMerkleProof,
		TransferIndex:       transferIndex,
		TransferTreeRoot:    transfersHash,
		Nonce:               nonce,
		TxTreeMerkleProof:   proposedBlock.TxTreeMerkleProof,
		TxIndex:             int32(txIndex),
		TxTreeRoot:          proposedBlock.TxTreeRoot,
	})
	if err != nil {
		log.Fatalf("failed to request withdrawal: %v", err)
	}

	log.Infof("The withdrawal request has been successfully sent.")
}

func ResumeWithdrawalRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	recipientAddressHex string,
	resumeIncompleteWithdrawals bool,
) {
	backupWithdrawals, err := GetBackupWithdrawal(ctx, cfg, log, common.HexToAddress(recipientAddressHex))
	if err != nil {
		log.Fatalf("failed to get backup withdrawal: %v", err)
	}

	var withdrawalInfo []*WithdrawalResponseData
	if len(backupWithdrawals) != 0 {
		const searchLimit = 10
		for i := 0; i < len(backupWithdrawals)/searchLimit; i += searchLimit {
			end := min(searchLimit*(i+1), len(backupWithdrawals))

			transferHashes := make([]string, end-searchLimit*i)
			for i, backupWithdrawal := range backupWithdrawals[searchLimit*i : end] {
				transferHashes[i] = hexutil.Encode(backupWithdrawal.Transfer.Hash().Marshal())
			}
			withdrawalInfo, err = FindWithdrawalsByTransferHashes(ctx, cfg, log, transferHashes)
			if err != nil {
				log.Fatalf("failed to find withdrawals: %v", err)
			}
		}
	}

	shouldProcess := func(withdrawal *tx_transfer_service.BackupWithdrawal) bool {
		transferHash := hexutil.Encode(withdrawal.Transfer.Hash().Marshal())
		for _, withdrawalInfo := range withdrawalInfo {
			if transferHash == withdrawalInfo.TransferHash {
				return false
			}
		}

		return true
	}

	incompleteBackupWithdrawals := make([]*tx_transfer_service.BackupWithdrawal, 0)
	for _, backupWithdrawal := range backupWithdrawals {
		if shouldProcess(backupWithdrawal) {
			incompleteBackupWithdrawals = append(incompleteBackupWithdrawals, backupWithdrawal)
		}
	}

	if len(incompleteBackupWithdrawals) == 0 {
		log.Infof("No incomplete withdrawal request found.")
		return
	}

	if !resumeIncompleteWithdrawals {
		log.Warnf("The withdrawal request has been found. Please use --resume flag to resume the withdrawal.")
		return
	}

	for _, backupWithdrawal := range incompleteBackupWithdrawals {
		// Send withdrawal request
		err = SendWithdrawalRequest(ctx, cfg, log, backupWithdrawal)
		if err != nil {
			log.Fatalf("failed to request withdrawal: %v", err)
		}

		log.Infof("The withdrawal request has been successfully sent.")
	}
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
	var withdrawal tx_transfer_service.BackupWithdrawal
	err = json.Unmarshal(encodedEncryptedTransfer, &withdrawal)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	// Send withdrawal transaction
	return SendWithdrawalRequest(ctx, cfg, log, &withdrawal)
}

func SendWithdrawalRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	withdrawal *tx_transfer_service.BackupWithdrawal,
) error {
	// Specify the block number containing the transaction.
	blockStatus, err := tx_transfer_service.GetBlockStatus(ctx, cfg, log, withdrawal.TransferTreeRoot)
	if err != nil {
		return fmt.Errorf("failed to get block status: %w", err)
	}
	blockNumber := blockStatus.BlockNumber

	rollupCfg := intMaxTypes.NewRollupContractConfigFromEnv(cfg, "https://sepolia-rpc.scroll.io")
	blockHash, err := intMaxTypes.FetchBlockHash(rollupCfg, ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch block hash: %w", err)
	}

	err = SendWithdrawalWithRawRequest(
		ctx, cfg, log,
		withdrawal.Transfer,
		withdrawal.TransferTreeRoot,
		withdrawal.Nonce,
		withdrawal.TransferMerkleProof,
		withdrawal.TransferIndex,
		withdrawal.TxTreeMerkleProof,
		withdrawal.TxIndex,
		blockNumber,
		blockHash,
	)
	if err != nil {
		return fmt.Errorf("failed to request withdrawal: %w", err)
	}

	return nil
}

func GetBackupWithdrawal(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	userAddress common.Address,
) ([]*tx_transfer_service.BackupWithdrawal, error) {
	userAllData, err := balance_service.GetUserBalancesRawRequest(ctx, cfg, lg, userAddress.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get user balances: %w", err)
	}

	withdrawals := make([]*tx_transfer_service.BackupWithdrawal, 0)
	for _, withdrawal := range userAllData.Transfers {
		// base64 decode
		var encodedEncryptedTransfer []byte
		encodedEncryptedTransfer, err = base64.StdEncoding.DecodeString(withdrawal.EncryptedTransfer)
		if err != nil {
			lg.Warnf("failed to decode base64: %w", err)
			continue
		}

		// json unmarshal
		var withdrawal tx_transfer_service.BackupWithdrawal
		err = json.Unmarshal(encodedEncryptedTransfer, &withdrawal)
		if err != nil {
			lg.Warnf("failed to unmarshal json: %w", err)
			continue
		}

		withdrawals = append(withdrawals, &withdrawal)
	}

	return withdrawals, nil
}
