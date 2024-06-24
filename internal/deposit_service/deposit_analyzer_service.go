package deposit_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const AMLRejectionThreshold = 70

func fetchNewDeposits(ctx context.Context, liquidity *bindings.Liquidity, startBlock uint64) ([]*bindings.LiquidityDeposited, uint64, map[uint32]bool, error) {
	nextBlock := startBlock + 1
	iterator, err := liquidity.FilterDeposited(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: ctx,
	}, [][32]byte{}, []uint64{}, []common.Address{})
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	var events []*bindings.LiquidityDeposited
	var maxLastSeenDepositIndex uint64
	tokenIndexMap := make(map[uint32]bool)

	for iterator.Next() {
		if iterator.Error() != nil {
			return nil, 0, nil, fmt.Errorf("error encountered: %w", iterator.Error())
		}
		event := iterator.Event
		events = append(events, event)
		tokenIndexMap[event.TokenIndex] = true

		if event.DepositIndex > maxLastSeenDepositIndex {
			maxLastSeenDepositIndex = event.DepositIndex
		}
	}

	return events, maxLastSeenDepositIndex, tokenIndexMap, nil
}

func getTokenInfoMap(ctx context.Context, liquidity *bindings.Liquidity, tokenIndexMap map[uint32]bool) (map[uint32]common.Address, error) {
	var tokenIndices []uint32
	for tokenIndex := range tokenIndexMap {
		tokenIndices = append(tokenIndices, tokenIndex)
	}

	tokenInfoMap := make(map[uint32]common.Address)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(tokenIndices))

	for _, tokenIndex := range tokenIndices {
		wg.Add(1)
		go func(tokenIndex uint32) {
			defer wg.Done()
			tokenInfo, err := liquidity.GetTokenInfo(&bind.CallOpts{
				Pending: false,
				Context: ctx,
			}, tokenIndex)
			if err != nil {
				errChan <- fmt.Errorf("failed to get token info for index %d: %w", tokenIndex, err)
				return
			}
			mu.Lock()
			tokenInfoMap[tokenIndex] = tokenInfo.TokenAddress
			mu.Unlock()
		}(tokenIndex)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan
	}

	return tokenInfoMap, nil
}

func fetchAMLScore(sender string, contractAddress string) uint32 {
	// TODO: Implement a real AML score fetching function
	return 50
}

func rejectDeposits(ctx context.Context, cfg *configs.Config, client *ethclient.Client, liquidity *bindings.Liquidity, maxLastSeenDepositIndex uint64, rejectedIndices []uint64) (*types.Receipt, error) {
	transactOpts, err := createTransactor(cfg)
	if err != nil {
		return nil, err
	}

	tx, err := liquidity.RejectDeposits(transactOpts, maxLastSeenDepositIndex, rejectedIndices)
	if err != nil {
		return nil, fmt.Errorf("failed to send RejectDeposits transaction: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func DepositAnalyzer(ctx context.Context, cfg *configs.Config, log logger.Logger) {
	client, err := newClient(cfg.Blockchain.EthreumNetworkRpcURL)
	if err != nil {
		log.Fatalf(err.Error())
	}

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		log.Fatalf("Failed to instantiate a Liquidity contract: %v", err.Error())
	}

	lastSeenDepositIndex, err := liquidity.GetLastSeenDepositIndex(&bind.CallOpts{
		Pending: false,
		Context: ctx,
	})
	if err != nil {
		log.Fatalf("Failed to get last seen deposit index: %v", err.Error())
	}

	lastSeenBlockNumber, err := fetchBlockNumberByDepositIndex(liquidity, lastSeenDepositIndex)
	if err != nil {
		log.Fatalf("Failed to get last block number: %v", err.Error())
	}

	events, maxLastSeenDepositIndex, tokenIndexMap, err := fetchNewDeposits(ctx, liquidity, lastSeenBlockNumber)
	if err != nil {
		log.Fatalf("Failed to fetch new deposits: %v", err.Error())
	}

	if len(events) == 0 {
		log.Infof("No new Deposited Events")
		return
	}

	tokenInfoMap, err := getTokenInfoMap(ctx, liquidity, tokenIndexMap)
	if err != nil {
		log.Fatalf("Failed to get token info map: %v", err.Error())
	}

	var rejectedIndices []uint64
	for _, event := range events {
		contractAddress := tokenInfoMap[event.TokenIndex]
		score := fetchAMLScore(event.Sender.Hex(), contractAddress.Hex())

		if score > AMLRejectionThreshold {
			rejectedIndices = append(rejectedIndices, event.DepositIndex)
		}
	}

	receipt, err := rejectDeposits(ctx, cfg, client, liquidity, maxLastSeenDepositIndex, rejectedIndices)
	if err != nil {
		log.Fatalf("Failed to reject deposits: %v", err.Error())
	}

	if receipt == nil {
		return
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("Successfully analyze %d deposits, %d rejects", len(events), len(rejectedIndices))
	} else {
		log.Infof("Failed to reject deposits")
	}

	log.Infof("Tx Hash: %v", receipt.TxHash.String())
}
