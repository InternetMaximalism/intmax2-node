//nolint:gocritic
package deposit_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"intmax2-node/pkg/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const fixedDepositValueInWei = 1e17 // 0.1 ETH in Wei

type DepositIndices struct {
	LastDepositAnalyzedEventInfo *DepositEventInfo
	LastDepositRelayedEventInfo  *DepositEventInfo
}

type DepositRelayerService struct {
	ctx       context.Context
	cfg       *configs.Config
	log       logger.Logger
	client    *ethclient.Client
	liquidity *bindings.Liquidity
}

func newDepositRelayerService(ctx context.Context, cfg *configs.Config, log logger.Logger, _ ServiceBlockchain) (*DepositRelayerService, error) {
	// link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
	// }

	client, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	return &DepositRelayerService{
		ctx:       ctx,
		cfg:       cfg,
		log:       log,
		client:    client,
		liquidity: liquidity,
	}, nil
}

// TODO: TxManager Class that stops processing if there are any pending transactions.
func DepositRelayer(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	depositRelayerService, err := newDepositRelayerService(ctx, cfg, log, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DepositRelayerService: %v", err.Error()))
	}

	_ = db.Exec(ctx, nil, func(d interface{}, input interface{}) error {
		q := d.(SQLDriverApp)

		blockNumberEvents, err := depositRelayerService.getBlockNumberEvents(q)
		if err != nil {
			panic(fmt.Sprintf("Failed to get block number events: %v", err.Error()))
		}

		depositIndices, err := depositRelayerService.fetchLastDepositEventIndices(
			blockNumberEvents[mDBApp.DepositsAnalyzedEvent].LastProcessedBlockNumber,
			blockNumberEvents[mDBApp.DepositsRelayedEvent].LastProcessedBlockNumber,
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to fetch deposit indices: %v", err.Error()))
		}

		if *depositIndices.LastDepositAnalyzedEventInfo.BlockNumber == uint64(0) {
			depositIndices.LastDepositAnalyzedEventInfo.BlockNumber = &cfg.Blockchain.LiquidityContractDeployedBlockNumber
		}

		if *depositIndices.LastDepositRelayedEventInfo.BlockNumber == uint64(0) {
			depositIndices.LastDepositRelayedEventInfo.BlockNumber = &cfg.Blockchain.LiquidityContractDeployedBlockNumber
		}

		_, err = q.UpsertEventBlockNumber(mDBApp.DepositsRelayedEvent, *depositIndices.LastDepositRelayedEventInfo.BlockNumber)
		if err != nil {
			panic(fmt.Sprintf("Error updating event block number: %v", err.Error()))
		}

		unprocessedDepositCount := *depositIndices.LastDepositAnalyzedEventInfo.LastDepositId - *depositIndices.LastDepositRelayedEventInfo.LastDepositId
		var shouldSubmit bool
		shouldSubmit, err = depositRelayerService.shouldProcessDepositRelayer(
			unprocessedDepositCount,
			*depositIndices.LastDepositRelayedEventInfo.BlockNumber,
		)
		if err != nil {
			panic(fmt.Sprintf("Error in threshold and time diff check: %v", err.Error()))
		}

		if !shouldSubmit {
			log.Infof(
				"Deposits will not be processed at this time. Unprocessed deposit count: %d, Last relayed block number: %d",
				unprocessedDepositCount,
				*depositIndices.LastDepositRelayedEventInfo.BlockNumber,
			)
			return nil
		}

		receipt, err := depositRelayerService.relayDeposits(*depositIndices.LastDepositAnalyzedEventInfo.LastDepositId, unprocessedDepositCount)
		if err != nil {
			panic(fmt.Sprintf("Failed to relay deposits: %v", err.Error()))
		}

		if receipt == nil {
			panic("Received nil receipt for transaction")
		}

		switch receipt.Status {
		case types.ReceiptStatusSuccessful:
			log.Infof("Successfully relay deposits. Transaction Hash: %v", receipt.TxHash.Hex())
		case types.ReceiptStatusFailed:
			panic(fmt.Sprintf("Transaction failed: relay deposits unsuccessful. Transaction Hash: %v", receipt.TxHash.Hex()))
		default:
			panic(fmt.Sprintf("Unexpected transaction status: %d. Transaction Hash: %v", receipt.Status, receipt.TxHash.Hex()))
		}

		return nil
	})
}

func (d *DepositRelayerService) getBlockNumberEvents(db SQLDriverApp) (map[string]*mDBApp.EventBlockNumber, error) {
	eventNames := []string{mDBApp.DepositsAnalyzedEvent, mDBApp.DepositsRelayedEvent}
	events, err := db.EventBlockNumbersByEventNames(eventNames)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event block numbers: %w", err)
	}

	results := make(map[string]*mDBApp.EventBlockNumber)
	for _, event := range events {
		results[event.EventName] = event
	}

	for _, eventName := range eventNames {
		if _, exists := results[eventName]; !exists {
			results[eventName] = &mDBApp.EventBlockNumber{
				EventName:                eventName,
				LastProcessedBlockNumber: d.cfg.Blockchain.LiquidityContractDeployedBlockNumber,
			}
		}
	}

	return results, nil
}

func (d *DepositRelayerService) fetchLastDepositEventIndices(depositAnalyzedBlockNumber, depositRelayedBlockNumber uint64) (DepositIndices, error) {
	type result struct {
		eventInfo *DepositEventInfo
		err       error
	}

	lastDepositAnalyzedIndexCh := make(chan result)
	lastDepositRelayedIndexCh := make(chan result)

	go func() {
		index, err := fetchLastDepositAnalyzedEvent(d.ctx, d.liquidity, depositAnalyzedBlockNumber)
		lastDepositAnalyzedIndexCh <- result{index, err}
	}()

	go func() {
		index, err := fetchLastDepositRelayedEvent(d.ctx, d.liquidity, depositRelayedBlockNumber)
		lastDepositRelayedIndexCh <- result{index, err}
	}()

	var di DepositIndices
	for {
		if di.LastDepositAnalyzedEventInfo != nil && di.LastDepositRelayedEventInfo != nil {
			return di, nil
		}
		select {
		case lastDepositAnalyzedResult := <-lastDepositAnalyzedIndexCh:
			if lastDepositAnalyzedResult.err != nil {
				return DepositIndices{}, lastDepositAnalyzedResult.err
			}
			di.LastDepositAnalyzedEventInfo = lastDepositAnalyzedResult.eventInfo
		case lastDepositRelayzedResult := <-lastDepositRelayedIndexCh:
			if lastDepositRelayzedResult.err != nil {
				return DepositIndices{}, lastDepositRelayzedResult.err
			}
			di.LastDepositRelayedEventInfo = lastDepositRelayzedResult.eventInfo
		}
	}
}

func (d *DepositRelayerService) shouldProcessDepositRelayer(unprocessedDepositCount, relayedBlockNumber uint64) (bool, error) {
	if unprocessedDepositCount <= 0 {
		return false, nil
	}

	if unprocessedDepositCount >= d.cfg.Blockchain.DepositRelayerThreshold {
		d.log.Infof("Deposit relayer threshold is reached. Unprocessed deposit count: %d", unprocessedDepositCount)
		return true, nil
	}

	depositIds := []*big.Int{}
	eventInfo, err := fetchDepositEvent(d.ctx, d.liquidity, relayedBlockNumber, depositIds)
	if err != nil {
		if err.Error() == "No deposit events found" {
			d.log.Infof("No deposit events found, skipping process")
			return false, nil
		}
		return false, fmt.Errorf("failed to get last block number: %w", err)
	}

	if *eventInfo.BlockNumber == 0 {
		return false, nil
	}

	isExceeded, err := isBlockTimeExceeded(d.ctx, d.client, *eventInfo.BlockNumber, int(d.cfg.Blockchain.DepositRelayerMinutesThreshold))
	if err != nil {
		return false, fmt.Errorf("error occurred while checking time difference: %w", err)
	}

	if !isExceeded {
		return false, nil
	}

	d.log.Infof("Block time difference exceeded the specified duration")
	return true, nil
}

func (d *DepositRelayerService) relayDeposits(maxLastSeenDepositIndex, numDepositsToRelay uint64) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(d.cfg.Blockchain.DepositRelayerPrivateKeyHex, d.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, err
	}

	transactOpts.Value = big.NewInt(fixedDepositValueInWei)
	upToDepositId := new(big.Int).SetUint64(maxLastSeenDepositIndex)
	gasLimit := new(big.Int).SetUint64(calculateRelayDepositsGasLimit(numDepositsToRelay))

	err = utils.LogTransactionDebugInfo(
		d.log,
		d.cfg.Blockchain.DepositRelayerPrivateKeyHex,
		d.cfg.Blockchain.LiquidityContractAddress,
		upToDepositId,
		gasLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to log transaction debug info: %w", err)
	}

	tx, err := d.liquidity.RelayDeposits(transactOpts, upToDepositId, gasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to send RelayDeposits transaction: %w", err)
	}

	receipt, err := bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func calculateRelayDepositsGasLimit(numDepositsToRelay uint64) uint64 {
	const (
		baseGas       = uint64(220000)
		perDepositGas = uint64(20000)
		bufferGas     = uint64(100000)
	)
	return baseGas + (perDepositGas * numDepositsToRelay) + bufferGas
}
