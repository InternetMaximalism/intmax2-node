package tx_withdrawal_service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/open_telemetry"
	intMaxTree "intmax2-node/internal/tree"
	"intmax2-node/internal/tx_transfer_service"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	withdrawalService "intmax2-node/internal/withdrawal_service"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	base10        = 10
	numUint32Bits = 32
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
		return fmt.Errorf("%s", errors.Join(ErrTokenNotFound, err))
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
	recipientBytes, err := hexutil.Decode(recipientAddressHex)
	if err != nil {
		return fmt.Errorf("failed to parse recipient address: %v", err)
	}

	recipientAddress, err := intMaxTypes.NewEthereumAddress(recipientBytes)
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
	const transferIndex = 0
	initialLeaves[transferIndex] = transfer

	transferTree, err := intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, initialLeaves, zeroTransfer.Hash())
	if err != nil {
		return fmt.Errorf("failed to create transfer tree: %v", err)
	}

	transferMerkleProof, transfersHash, err := transferTree.ComputeMerkleProof(0)
	if err != nil {
		return fmt.Errorf("failed to compute merkle proof: %v", err)
	}

	var nonce uint32 = 1 // TODO: Incremented with each transaction

	err = SendWithdrawalTransaction(
		ctx,
		cfg,
		log,
		sb,
		userAccount,
		transfersHash,
		nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Println("The transaction request has been successfully sent. Please wait for the server's response.")

	// Get proposed block
	proposedBlock, err := tx_transfer_service.GetBlockProposed(
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

	publicKeysStr := make([]string, len(proposedBlock.PublicKeys))
	for i, key := range proposedBlock.PublicKeys {
		publicKeysStr[i] = key.ToAddress().String()
	}

	txIndex := slices.Index(publicKeysStr, userAccount.ToAddress().String())
	if txIndex == -1 {
		return fmt.Errorf("failed to find user's public key in the proposed block")
	}

	txDetails := intMaxTypes.TxDetails{
		Tx: intMaxTypes.Tx{
			TransferTreeRoot: &transfersHash,
			Nonce:            nonce,
		},
		TxTreeRoot:    &proposedBlock.TxTreeRoot,
		TxMerkleProof: proposedBlock.TxTreeMerkleProof,
		Transfers:     initialLeaves,
	}
	backupTx, err := transaction.NewBackupTransactionData(
		userAccount.Public(),
		txDetails,
		txHash,
		"0x",
	)
	if err != nil {
		return fmt.Errorf("failed to create backup transaction data: %w", err)
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
			return fmt.Errorf("failed to make backup data: %v", err)
		}
	}

	// Accept proposed block
	err = tx_transfer_service.SendSignedProposedBlock(
		ctx, cfg, userAccount, proposedBlock.TxTreeRoot, *txHash, proposedBlock.PublicKeys,
		backupTx, backupTransfers,
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	fmt.Println("The transaction has been successfully sent.")

	// Send withdrawal request
	err = SendWithdrawalRequest(ctx, cfg, log, sb, &tx_transfer_service.BackupWithdrawal{
		SenderAddress:       userAccount.ToAddress(),
		Transfer:            transfer,
		TransferMerkleProof: transferMerkleProof,
		TransferIndex:       transferIndex,
		TransferTreeRoot:    transfersHash,
		Nonce:               uint64(nonce),
		TxTreeMerkleProof:   proposedBlock.TxTreeMerkleProof,
		TxIndex:             int32(txIndex),
		TxTreeRoot:          proposedBlock.TxTreeRoot,
	})
	if err != nil {
		return fmt.Errorf("failed to request withdrawal: %v", err)
	}

	fmt.Println("The withdrawal request has been successfully sent.")

	return nil
}

func ResumeWithdrawalRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	recipientAddressHex string,
	resumeIncompleteWithdrawals bool,
) {
	backupWithdrawals, err := GetBackupWithdrawal(ctx, cfg, common.HexToAddress(recipientAddressHex))
	if err != nil {
		log.Fatalf("failed to get backup withdrawal: %v", err)
	}

	var withdrawalInfo []*WithdrawalResponseData
	if len(backupWithdrawals) != 0 {
		const searchLimit = 10
		numChunks := (len(backupWithdrawals) + searchLimit - 1) / searchLimit
		for i := 0; i < numChunks; i += searchLimit {
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
		for _, wi := range withdrawalInfo {
			if transferHash == wi.TransferHash {
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
		log.Debugf("No incomplete withdrawal request found.")
		return
	}

	if !resumeIncompleteWithdrawals {
		log.Fatalf("The withdrawal request has been found. Please use --resume flag to resume the withdrawal.")
		return
	}

	for _, backupWithdrawal := range incompleteBackupWithdrawals {
		// Send withdrawal request
		err = SendWithdrawalRequest(ctx, cfg, log, sb, backupWithdrawal)
		if err != nil {
			if errors.Is(err, withdrawalService.ErrWithdrawalRequestAlreadyExists) {
				log.Warnf("The withdrawal request already exists.")
				continue
			}

			if errors.Is(err, tx_transfer_service.ErrBlockNotFound) {
				log.Warnf("The block containing the transaction is not posted yet.")
				continue
			}

			log.Fatalf("failed to request withdrawal: %v", err)
		}

		fmt.Println("The withdrawal request has been successfully sent.")
	}
}

func SendWithdrawalTransactionFromBackupTransfer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
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
	return SendWithdrawalRequest(ctx, cfg, log, sb, &withdrawal)
}

func SendWithdrawalRequest(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	withdrawal *tx_transfer_service.BackupWithdrawal,
) error {
	err := sb.SetupScrollNetworkChainID(ctx)
	if err != nil {
		open_telemetry.MarkSpanError(ctx, err)
		return errors.Join(errorsB.ErrSetupScrollNetworkChainIDFail, err)
	}

	// Specify the block number containing the transaction.
	blockStatus, err := tx_transfer_service.GetBlockStatus(ctx, cfg, log, withdrawal.TxTreeRoot)
	if err != nil {
		if errors.Is(err, ErrBlockNotFound) {
			return ErrBlockNotFound
		}
		return fmt.Errorf("failed to get block status: %w", err)
	}

	var link string
	link, err = sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		open_telemetry.MarkSpanError(ctx, err)
		return errors.Join(errorsB.ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	rollupCfg := intMaxTypes.NewRollupContractConfigFromEnv(cfg, link)
	blockNumber, err := strconv.ParseUint(blockStatus.BlockNumber, base10, numUint32Bits)
	if err != nil {
		return fmt.Errorf("failed to parse block number: %w", err)
	}
	blockHash, err := intMaxTypes.FetchBlockHash(rollupCfg, ctx, uint32(blockNumber))
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
		uint32(blockNumber),
		blockHash,
	)
	if err != nil {
		if errors.Is(err, withdrawalService.ErrWithdrawalRequestAlreadyExists) {
			return withdrawalService.ErrWithdrawalRequestAlreadyExists
		}

		return fmt.Errorf("failed to request withdrawal: %w", err)
	}

	return nil
}

func GetBackupWithdrawal(
	ctx context.Context,
	cfg *configs.Config,
	userAddress common.Address,
) ([]*tx_transfer_service.BackupWithdrawal, error) {
	userAllData, err := balance_service.GetUserBalancesRawRequest(ctx, cfg, userAddress.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to get user balances: %w", err)
	}

	withdrawals := make([]*tx_transfer_service.BackupWithdrawal, 0)
	for _, withdrawal := range userAllData.Transfers {
		// base64 decode
		var encodedEncryptedTransfer []byte
		encodedEncryptedTransfer, err = base64.StdEncoding.DecodeString(withdrawal.EncryptedTransfer)
		if err != nil {
			fmt.Printf("Warning: failed to decode base64: %v", err)
			continue
		}

		// json unmarshal
		var withdrawal tx_transfer_service.BackupWithdrawal
		err = json.Unmarshal(encodedEncryptedTransfer, &withdrawal)
		if err != nil {
			fmt.Printf("Warning: failed to unmarshal json: %v", err)
			continue
		}

		withdrawals = append(withdrawals, &withdrawal)
	}

	return withdrawals, nil
}
