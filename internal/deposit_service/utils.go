package deposit_service

import (
	"context"
	"fmt"
	"intmax2-node/internal/bindings"
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

func fetchLastDepositAnalyzedEvent(ctx context.Context, liquidity *bindings.Liquidity, startBlockNumber uint64) (*DepositEventInfo, error) {
	iterator, err := liquidity.FilterDepositsAnalyzed(&bind.FilterOpts{
		Start:   startBlockNumber,
		End:     nil,
		Context: ctx,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}

	defer func() {
		_ = iterator.Close()
	}()

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

func fetchLastDepositRelayedEvent(ctx context.Context, liquidity *bindings.Liquidity, startBlockNumber uint64) (*DepositEventInfo, error) {
	iterator, err := liquidity.FilterDepositsRelayed(&bind.FilterOpts{
		Start:   startBlockNumber,
		End:     nil,
		Context: ctx,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}
	defer func() {
		_ = iterator.Close()
	}()

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

func fetchDepositEvent(ctx context.Context, liquidity *bindings.Liquidity, startBlockNumber uint64, depositIds []*big.Int) (*DepositEventInfo, error) {
	nextBlock := startBlockNumber + 1
	iterator, err := liquidity.FilterDeposited(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: ctx,
	}, depositIds, []common.Address{})
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}
	defer func() {
		_ = iterator.Close()
	}()

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

func isBlockTimeExceeded(ctx context.Context, client *ethclient.Client, blockNumber uint64, minutes int) (bool, error) {
	block, err := client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
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
