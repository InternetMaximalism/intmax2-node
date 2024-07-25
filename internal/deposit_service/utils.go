package deposit_service

import (
	"fmt"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
		return nil, fmt.Errorf("no deposits relayed events found")
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
		return nil, fmt.Errorf("no deposits relayed events found")
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

func updateEventBlockNumber(db SQLDriverApp, log logger.Logger, eventName string, blockNumber int64) error {
	updatedEvent, err := db.UpsertEventBlockNumber(eventName, blockNumber)
	if err != nil {
		return err
	}
	log.Infof("Updated %s block number to %d", eventName, updatedEvent.LastProcessedBlockNumber)
	return nil
}
