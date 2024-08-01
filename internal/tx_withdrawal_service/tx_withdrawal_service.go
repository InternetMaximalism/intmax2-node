package tx_transfer_service

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	"intmax2-node/internal/tx_transfer_service"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func SendWithdrawalTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
	args []string,
	amountStr string,
	recipientAddressHex string,
	userPrivateKey string,
) {
	userAccount, err := intMaxAcc.NewPrivateKeyFromString(userPrivateKey)
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

	balance, err := balance_service.GetUserBalance(ctx, cfg, log, db, userAccount, tokenIndex)
	if err != nil {
		log.Fatalf(ErrFailedToGetBalance+": %v", err)
	}

	if strings.TrimSpace(amountStr) == "" {
		log.Fatalf("Amount is required")
	}

	const int10Key = 10
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		log.Fatalf("failed to convert amount to int: %v", err)
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
	initialLeaves[0] = transfer

	transferTree, err := intMaxTree.NewTransferTree(intMaxTree.TRANSFER_TREE_HEIGHT, initialLeaves, zeroTransfer.Hash())
	if err != nil {
		log.Fatalf("failed to create transfer tree: %v", err)
	}

	transfersHash, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	var nonce uint64 = 1 // TODO: Incremented with each transaction
	err = SendWithdrawalRequest(
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

	log.Printf("The proposed block has been successfully received. Please wait for the server's response.")

	tx, err := intMaxTypes.NewTx(
		&transfersHash,
		nonce,
	)
	if err != nil {
		log.Fatalf("failed to create new tx: %w", err)
	}

	txHash := tx.Hash()

	// Accept proposed block
	err = tx_transfer_service.SendSignedProposedBlock(
		ctx, cfg, log, userAccount, proposedBlock.TxTreeRoot, *txHash, proposedBlock.PublicKeys,
	)
	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	}

	log.Printf("The transaction has been successfully sent.")

	// TODO: Get the block number and block hash
	blockNumber := uint32(0)
	blockHash := common.Hash{}.Hex()

	err = SendWithdrawalWithRawRequest(
		ctx, cfg, log, userAccount, transfer, transfersHash, nonce, blockNumber, blockHash,
	)
	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	}

	log.Printf("The transaction has been successfully sent.")
}
