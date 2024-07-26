package tx_transfer_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"os"
	"strings"
)

func SendTransferTransaction(
	ctx context.Context,
	cfg *configs.Config,
	lg logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
	args []string,
	amountStr string,
	recipientAddressStr string,
	userPrivateKey string,
) {
	userAccount, err := intMaxAcc.NewPrivateKeyFromString(userPrivateKey)
	if err != nil {
		fmt.Printf("fail to parse user private key: %v\n", err)
		os.Exit(1)
	}

	tokenInfo, err := new(intMaxTypes.TokenInfo).ParseFromStrings(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tokenIndex, err := balance_service.GetTokenIndex(ctx, cfg, db, sb, *tokenInfo)
	if err != nil {
		fmt.Println(ErrTokenNotFound, err)
		os.Exit(1)
	}

	balance, err := balance_service.GetUserBalance(db, userAccount.ToAddress(), tokenIndex)
	if err != nil {
		fmt.Printf(ErrFailedToGetBalance+": %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(amountStr) == "" {
		fmt.Println("Amount is required")
		os.Exit(1)
	}

	const int10Key = 10
	amount, ok := new(big.Int).SetString(amountStr, int10Key)
	if !ok {
		fmt.Printf("failed to convert amount to int: %v\n", err)
		os.Exit(1)
	}

	if balance.Cmp(amount) < 0 {
		fmt.Printf("Insufficient balance: %s\n", balance)
		os.Exit(1)
	}

	// Send transfer transaction
	recipient, err := intMaxAcc.NewPublicKeyFromAddressHex(recipientAddressStr)
	if err != nil {
		fmt.Printf("failed to parse recipient address: %v\n", err)
		os.Exit(1)
	}

	recipientAddress, err := intMaxTypes.NewINTMAXAddress(recipient.ToAddress().Bytes())
	if err != nil {
		fmt.Printf("failed to create recipient address: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("failed to create transfer tree: %v\n", err)
		os.Exit(1)
	}

	transfersHash, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	var nonce uint64 = 1 // TODO: Incremented with each transaction
	err = SendTransactionRequest(
		cfg,
		ctx,
		userAccount,
		transfersHash,
		nonce,
	)
	if err != nil {
		fmt.Printf("failed to send transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("The transaction request has been successfully sent. Please wait for the server's response.")

	// Get proposed block
	proposedBlock, err := GetBlockProposed(
		ctx, userAccount, transfersHash, nonce,
	)
	if err != nil {
		fmt.Printf("failed to send transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("The proposed block has been successfully received. Please wait for the server's response.")

	// Accept proposed block
	err = SendSignedProposedBlock(
		userAccount, proposedBlock.TxTreeRoot, proposedBlock.PublicKeysHash,
	)
	if err != nil {
		fmt.Printf("failed to send transaction: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("The transaction has been successfully sent.")
}
