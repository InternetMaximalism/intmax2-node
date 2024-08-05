package tx_claim_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/pkg/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ClaimWithdrawals(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	userEthPrivateKey string,
) {
	privateKey, err := crypto.HexToECDSA(userEthPrivateKey)
	if err != nil {
		log.Fatalf("Failed to convert hex to ECDSA: %v", err)
	}

	// アドレスを取得
	recipientEthAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	withdrawals, err := fetchRecipientClaimableWithdrawals(ctx, cfg, sb, recipientEthAddress)
	if err != nil {
		log.Errorf("Failed to fetch recipient claimable withdrawals: %v", err)
	}

	receipts, err := claimWithdrawals(ctx, cfg, log, sb, withdrawals, userEthPrivateKey)
	if err != nil {
		log.Errorf("Failed to claim withdrawals: %v", err)
	}

	if len(receipts) != 0 {
		log.Infof("The claiming withdrawals has been successfully sent.")
	}
}

var ErrScrollNetworkChainLinkEvmJSONRPCFail = errors.New("failed to get Scroll network chain link")
var ErrCreateNewClientOfRPCEthFail = errors.New("failed to create new client of RPC Ethereum")
var ErrNewBlockBuilderRegistryCallerFail = errors.New("failed to create new the block builder registry caller")

// fetch all ClaimableWithdrawalQueued events in Withdrawal contract and filter by recipient address
func fetchRecipientClaimableWithdrawals(
	ctx context.Context,
	cfg *configs.Config,
	sb ServiceBlockchain,
	recipientEthAddress common.Address,
) ([]bindings.WithdrawalLibWithdrawal, error) {
	const searchBlocksLimitAtOnce uint64 = 5000
	const (
		hName = "TxClaimService func:fetchRecipientClaimableWithdrawals"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	link, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = ethclient.Dial(link)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrCreateNewClientOfRPCEthFail, err)
	}
	defer func() {
		client.Close()
	}()

	withdrawalContract, err := bindings.NewWithdrawal(
		common.HexToAddress(cfg.Blockchain.WithdrawalContractAddress),
		client,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrNewBlockBuilderRegistryCallerFail, err)
	}

	latestBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, err
	}

	// fetch all ClaimableWithdrawalQueued events in Withdrawal contract
	// and filter by recipient address
	// 5000 is the maximum number of blocks to fetch

	startBlock := cfg.Blockchain.RollupContractDeployedBlockNumber

	withdrawals := make([]bindings.WithdrawalLibWithdrawal, 0)
	for {
		opts := bind.FilterOpts{
			Start: startBlock,
		}
		if startBlock+searchBlocksLimitAtOnce <= latestBlockNumber {
			endBlock := startBlock + searchBlocksLimitAtOnce
			opts.End = &endBlock
		}
		var events *bindings.WithdrawalClaimableWithdrawalQueuedIterator
		events, err = withdrawalContract.FilterClaimableWithdrawalQueued(&opts)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return nil, err
		}

		for events.Next() {
			// filter by recipient address
			if events.Event.Withdrawal.Recipient == recipientEthAddress {
				withdrawals = append(withdrawals, events.Event.Withdrawal)
			}
		}

		if opts.End == nil {
			// Searched all blocks
			break
		}

		startBlock = *opts.End + 1
	}

	return withdrawals, nil
}

// fetch all ClaimableWithdrawalQueued events in Withdrawal contract and filter by recipient address
func claimWithdrawals(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	withdrawals []bindings.WithdrawalLibWithdrawal,
	userEthPrivateKey string,
) ([]*types.Receipt, error) {
	const (
		hName = "TxClaimService func:claimWithdrawals"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	var client *ethclient.Client
	client, err = ethclient.Dial(link)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrCreateNewClientOfRPCEthFail, err)
	}
	defer func() {
		client.Close()
	}()

	liquidityContract, err := bindings.NewLiquidity(
		common.HexToAddress(cfg.Blockchain.WithdrawalContractAddress),
		client,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrNewBlockBuilderRegistryCallerFail, err)
	}

	transactOpts, err := utils.CreateTransactor(userEthPrivateKey, cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	receipts := make([]*types.Receipt, 0)
	count := 0
	for i := range withdrawals {
		var withdrawalJSON []byte
		withdrawalJSON, err = json.Marshal(&withdrawals[i])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal withdrawal: %w", err)
		}
		log.Debugf("Claiming withdrawal[%d]: %s\n", i, withdrawalJSON)
		var tx *types.Transaction
		tx, err = liquidityContract.ClaimWithdrawals(transactOpts, []bindings.WithdrawalLibWithdrawal{withdrawals[i]})
		if err != nil {
			// TODO: Continue only if WithdrawalNotFound error was occurred.
			// return nil, fmt.Errorf("failed to claim withdrawals: %w", err)
			log.Warnf("Failed to claim withdrawals[%d]: %v", i, err)
			continue
		}

		log.Infof("ClaimWithdrawals tx sent: %s", tx.Hash().Hex())

		receipt, err := bind.WaitMined(ctx, client, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
		}

		receipts = append(receipts, receipt)
		count += 1
	}

	log.Infof("Claimed %d withdrawals", count)

	return receipts, nil
}
