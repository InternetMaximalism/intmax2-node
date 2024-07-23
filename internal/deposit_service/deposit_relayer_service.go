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

type DepositIndices struct {
	LastDepositAnalyzedEventInfo *DepositEventInfo
	LastDepositRelayedEventInfo  *DepositEventInfo
}

func getBlockNumberEvents(db SQLDriverApp) (map[string]*mDBApp.EventBlockNumber, error) {
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
				LastProcessedBlockNumber: 0,
			}
		}
	}

	return results, nil
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

func fetchLastDepositEventIndices(liquidity *bindings.Liquidity, depositAnalyzedBlockNumber uint64, depositRelayedBlockNumber uint64) (DepositIndices, error) {
	type result struct {
		eventInfo *DepositEventInfo
		err       error
	}

	lastDepositAnalyzedIndexCh := make(chan result)
	lastDepositRelayedIndexCh := make(chan result)

	go func() {
		index, err := fetchLastDepositAnalyzedEvent(liquidity, depositAnalyzedBlockNumber)
		lastDepositAnalyzedIndexCh <- result{index, err}
	}()

	go func() {
		index, err := fetchLastDepositRelayedEvent(liquidity, depositRelayedBlockNumber)
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

func shouldProcessDeposits(client *ethclient.Client, liquidity *bindings.Liquidity, unprocessedDepositCount, relayedBlockNumber uint64) (bool, error) {
	if unprocessedDepositCount <= 0 {
		return false, nil
	}

	if unprocessedDepositCount >= depositThreshold {
		fmt.Println("Deposit threshold is reached")
		return true, nil
	}

	depositIds := []*big.Int{}
	eventInfo, err := fetchDepositEvent(liquidity, relayedBlockNumber, depositIds)
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

	isExceeded, err := isBlockTimeExceeded(client, *eventInfo.BlockNumber)
	if err != nil {
		return false, fmt.Errorf("error occurred while checking time difference: %w", err)
	}

	if !isExceeded {
		return false, nil
	}

	fmt.Println("Block time difference exceeded the specified duration")
	return true, nil
}

func calculateRelayDepositsGasLimit(numDepositsToRelay uint64) uint64 {
	const (
		baseGas       = uint64(220000)
		perDepositGas = uint64(20000)
		bufferGas     = uint64(100000)
	)
	return baseGas + (perDepositGas * numDepositsToRelay) + bufferGas
}

func relayDeposits(ctx context.Context, cfg *configs.Config, client *ethclient.Client, liquidity *bindings.Liquidity, maxLastSeenDepositIndex uint64, numDepositsToRelay uint64) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(cfg)
	if err != nil {
		return nil, err
	}

	transactOpts.Value = big.NewInt(0).Mul(big.NewInt(1e17), big.NewInt(1)) // NOTE: Fixed Value 0.1 ETH
	gasLimit := calculateRelayDepositsGasLimit(numDepositsToRelay)

	tx, err := liquidity.RelayDeposits(transactOpts, new(big.Int).SetUint64(uint64(maxLastSeenDepositIndex)), new(big.Int).SetUint64(uint64(gasLimit)))
	if err != nil {
		return nil, fmt.Errorf("failed to send RelayDeposits transaction: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func DepositRelayer(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var client *ethclient.Client
	client, err = utils.NewClient(link)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer client.Close()

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		log.Fatalf("Failed to instantiate a Liquidity contract: %v", err.Error())
	}

	blockNumberEvents, err := getBlockNumberEvents(db)
	if err != nil {
		log.Errorf("Failed to get block number events: %v", err)
		return
	}

	depositIndices, err := fetchLastDepositEventIndices(
		liquidity,
		uint64(blockNumberEvents[mDBApp.DepositsAnalyzedEvent].LastProcessedBlockNumber),
		uint64(blockNumberEvents[mDBApp.DepositsRelayedEvent].LastProcessedBlockNumber),
	)
	if err != nil {
		log.Fatalf("Failed to fetch deposit indices: %v", err.Error())
	}

	unprocessedDepositCount := *depositIndices.LastDepositAnalyzedEventInfo.LastDepositId - *depositIndices.LastDepositRelayedEventInfo.LastDepositId
	shouldSubmit, err := shouldProcessDeposits(
		client,
		liquidity,
		unprocessedDepositCount,
		*depositIndices.LastDepositRelayedEventInfo.BlockNumber,
	)
	if err != nil {
		log.Fatalf("Error in threshold and time diff check: %v", err)
		return
	}

	if !shouldSubmit {
		return
	}

	receipt, err := relayDeposits(ctx, cfg, client, liquidity, *depositIndices.LastDepositAnalyzedEventInfo.LastDepositId, unprocessedDepositCount)
	if err != nil {
		log.Fatalf("Failed to relay deposits: %v", err.Error())
	}

	if receipt == nil {
		return
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("Successfully relay deposits")
	} else {
		log.Infof("Failed to relay deposits")
	}

	switch receipt.Status {
	case types.ReceiptStatusSuccessful:
		log.Infof("Successfully relay deposits")
	case types.ReceiptStatusFailed:
		log.Errorf("Transaction failed: relay deposits unsuccessful")
	default:
		log.Warnf("Unexpected transaction status: %d", receipt.Status)
	}

	log.Infof("Transaction hash: %s", receipt.TxHash.Hex())

	updatedEvent, err := db.UpsertEventBlockNumber(mDBApp.DepositsRelayedEvent, int64(*depositIndices.LastDepositRelayedEventInfo.BlockNumber))
	if err != nil {
		log.Errorf("Failed to upsert event block number: %v", err)
		return
	}
	log.Infof("Updated DepositsRelayedEvent block number to %d", updatedEvent.LastProcessedBlockNumber)
}
