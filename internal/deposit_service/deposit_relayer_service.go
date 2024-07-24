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
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const depositThreshold uint64 = 128
const duration = 1 * time.Hour
const FixedDepositValueInWei = 1e17 // 0.1 ETH in Wei

type DepositIndices struct {
	LastDepositAnalyzedEventInfo *DepositEventInfo
	LastDepositRelayedEventInfo  *DepositEventInfo
}

type DepositRelayerService struct {
	ctx       context.Context
	cfg       *configs.Config
	log       logger.Logger
	db        SQLDriverApp
	client    *ethclient.Client
	liquidity *bindings.Liquidity
}

func newDepositRelayerService(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, _ ServiceBlockchain) (*DepositRelayerService, error) {
	// link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
	// }

	client, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	return &DepositRelayerService{
		ctx:       ctx,
		cfg:       cfg,
		log:       log,
		db:        db,
		client:    client,
		liquidity: liquidity,
	}, nil
}

// TODO: TxManager Class that stops processing if there are any pending transactions.
func DepositRelayer(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	depositRelayerService, err := newDepositRelayerService(ctx, cfg, log, db, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DepositRelayerService: %v", err.Error()))
	}

	blockNumberEvents, err := depositRelayerService.getBlockNumberEvents()
	if err != nil {
		panic(fmt.Sprintf("Failed to get block number events: %v", err.Error()))
	}

	depositIndices, err := depositRelayerService.fetchLastDepositEventIndices(
		uint64(blockNumberEvents[mDBApp.DepositsAnalyzedEvent].LastProcessedBlockNumber),
		uint64(blockNumberEvents[mDBApp.DepositsRelayedEvent].LastProcessedBlockNumber),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch deposit indices: %v", err.Error()))
	}

	unprocessedDepositCount := *depositIndices.LastDepositAnalyzedEventInfo.LastDepositId - *depositIndices.LastDepositRelayedEventInfo.LastDepositId
	shouldSubmit, err := depositRelayerService.shouldProcessDeposits(
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
		return
	}

	receipt, err := depositRelayerService.relayDeposits(*depositIndices.LastDepositAnalyzedEventInfo.LastDepositId, unprocessedDepositCount)
	if err != nil {
		panic(fmt.Sprintf("Failed to relay deposits: %v", err.Error()))
	}

	if receipt == nil {
		return
	}

	switch receipt.Status {
	case types.ReceiptStatusSuccessful:
		log.Infof("Successfully relay deposits")
	case types.ReceiptStatusFailed:
		panic("Transaction failed: relay deposits unsuccessful")
	default:
		log.Warnf("Unexpected transaction status: %d", receipt.Status)
	}

	log.Infof("Transaction hash: %s", receipt.TxHash.Hex())

	updateEventBlockNumber(db, log, mDBApp.DepositsRelayedEvent, int64(*depositIndices.LastDepositRelayedEventInfo.BlockNumber))
}

func (d *DepositRelayerService) getBlockNumberEvents() (map[string]*mDBApp.EventBlockNumber, error) {
	eventNames := []string{mDBApp.DepositsAnalyzedEvent, mDBApp.DepositsRelayedEvent}
	events, err := d.db.EventBlockNumbersByEventNames(eventNames)
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
				LastProcessedBlockNumber: 0,
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
		index, err := fetchLastDepositAnalyzedEvent(d.liquidity, depositAnalyzedBlockNumber)
		lastDepositAnalyzedIndexCh <- result{index, err}
	}()

	go func() {
		index, err := fetchLastDepositRelayedEvent(d.liquidity, depositRelayedBlockNumber)
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

func (d *DepositRelayerService) shouldProcessDeposits(unprocessedDepositCount, relayedBlockNumber uint64) (bool, error) {
	if unprocessedDepositCount <= 0 {
		return false, nil
	}

	if unprocessedDepositCount >= depositThreshold {
		fmt.Println("Deposit threshold is reached")
		return true, nil
	}

	depositIds := []*big.Int{}
	eventInfo, err := fetchDepositEvent(d.liquidity, relayedBlockNumber, depositIds)
	if err != nil {
		if err.Error() == "No deposit events found" {
			fmt.Println("No deposit events found, skipping process")
			return false, nil
		}
		return false, fmt.Errorf("failed to get last block number: %w", err)
	}

	if *eventInfo.BlockNumber == 0 {
		return false, nil
	}

	isExceeded, err := isBlockTimeExceeded(d.client, *eventInfo.BlockNumber)
	if err != nil {
		return false, fmt.Errorf("error occurred while checking time difference: %w", err)
	}

	if !isExceeded {
		return false, nil
	}

	fmt.Println("Block time difference exceeded the specified duration")
	return true, nil
}

func (d *DepositRelayerService) relayDeposits(maxLastSeenDepositIndex, numDepositsToRelay uint64) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(d.cfg)
	if err != nil {
		return nil, err
	}

	transactOpts.Value = big.NewInt(FixedDepositValueInWei)
	gasLimit := calculateRelayDepositsGasLimit(numDepositsToRelay)

	tx, err := d.liquidity.RelayDeposits(transactOpts, new(big.Int).SetUint64(maxLastSeenDepositIndex), new(big.Int).SetUint64(gasLimit))
	if err != nil {
		return nil, fmt.Errorf("failed to send RelayDeposits transaction: %w", err)
	}

	receipt, err := bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func isBlockTimeExceeded(client *ethclient.Client, blockNumber uint64) (bool, error) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		return false, fmt.Errorf("failed to get block by number: %w", err)
	}

	timestamp := block.Time()
	currentTime := time.Now()

	blockTime := time.Unix(int64(timestamp), 0)
	diff := currentTime.Sub(blockTime)

	return diff > duration, nil
}

func calculateRelayDepositsGasLimit(numDepositsToRelay uint64) uint64 {
	const (
		baseGas       = uint64(220000)
		perDepositGas = uint64(20000)
		bufferGas     = uint64(100000)
	)
	return baseGas + (perDepositGas * numDepositsToRelay) + bufferGas
}
