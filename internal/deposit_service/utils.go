package deposit_service

import (
	"context"
	"fmt"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type DepositEventInfo struct {
	LastDepositId *uint64
	BlockNumber   *uint64
}

func fetchLastDepositAnalyzedEvent(liquidity *bindings.Liquidity, startBlockNumber uint64) (*DepositEventInfo, error) {
	iterator, err := liquidity.FilterDepositsAnalyzed(&bind.FilterOpts{
		Start: startBlockNumber,
		End:   nil,
	}, []*big.Int{})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}

	defer iterator.Close()

	var lastEvent *DepositEventInfo

	for iterator.Next() {
		if iterator.Error() != nil {
			return nil, fmt.Errorf("error encountered while iterating: %v", iterator.Error())
		}

		currentId := iterator.Event.LastAnalyzedDepositId.Uint64()
		currentBlockNumber := iterator.Event.Raw.BlockNumber

		lastEvent = &DepositEventInfo{
			LastDepositId: &currentId,
			BlockNumber:   &currentBlockNumber,
		}
	}

	if lastEvent == nil {
		lastDepositId := uint64(0)
		blockNumber := uint64(0)
		return &DepositEventInfo{
			LastDepositId: &lastDepositId,
			BlockNumber:   &blockNumber,
		}, nil
	}

	return lastEvent, nil
}

func fetchLastDepositRelayedEvent(liquidity *bindings.Liquidity, startBlockNumber uint64) (*DepositEventInfo, error) {
	iterator, err := liquidity.FilterDepositsRelayed(&bind.FilterOpts{
		Start: startBlockNumber,
		End:   nil,
	}, []*big.Int{})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}
	defer iterator.Close()

	var lastEvent *DepositEventInfo

	for iterator.Next() {
		if iterator.Error() != nil {
			return nil, fmt.Errorf("error encountered while iterating: %v", iterator.Error())
		}

		currentId := iterator.Event.LastRelayedDepositId.Uint64()
		currentBlockNumber := iterator.Event.Raw.BlockNumber

		lastEvent = &DepositEventInfo{
			LastDepositId: &currentId,
			BlockNumber:   &currentBlockNumber,
		}
	}

	if lastEvent == nil {
		lastDepositId := uint64(0)
		blockNumber := uint64(0)
		return &DepositEventInfo{
			LastDepositId: &lastDepositId,
			BlockNumber:   &blockNumber,
		}, nil
	}

	return lastEvent, nil
}

func fetchDepositEvent(liquidity *bindings.Liquidity, startBlockNumber uint64, depositIds []*big.Int) (*DepositEventInfo, error) {
	iterator, err := liquidity.FilterDeposited(&bind.FilterOpts{
		Start: startBlockNumber,
		End:   nil,
	}, depositIds, []common.Address{})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}
	defer iterator.Close()

	var event *DepositEventInfo

	for iterator.Next() {
		if iterator.Error() != nil {
			return nil, fmt.Errorf("error encountered while iterating: %v", iterator.Error())
		}

		currentId := iterator.Event.DepositId.Uint64()
		currentBlockNumber := iterator.Event.Raw.BlockNumber

		event = &DepositEventInfo{
			LastDepositId: &currentId,
			BlockNumber:   &currentBlockNumber,
		}
	}

	if event == nil {
		return nil, fmt.Errorf("no deposit events found")
	}

	return event, nil
}

func updateEventBlockNumber(db SQLDriverApp, log logger.Logger, eventName string, blockNumber uint64) error {
	updatedEvent, err := db.UpsertEventBlockNumber(eventName, blockNumber)
	if err != nil {
		return err
	}
	log.Infof("Updated %s block number to %d", eventName, updatedEvent.LastProcessedBlockNumber)
	return nil
}

func isBlockTimeExceeded(client *ethclient.Client, blockNumber uint64, minutes int) (bool, error) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		return false, fmt.Errorf("failed to get block by number: %w", err)
	}

	timestamp := block.Time()
	currentTime := time.Now()

	blockTime := time.Unix(int64(timestamp), 0)
	diff := currentTime.Sub(blockTime)
	duration := time.Duration(minutes) * time.Minute

	return diff > duration, nil
}
