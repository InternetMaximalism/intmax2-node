package deposit_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"math/big"
	"sync"

	"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const AMLRejectionThreshold = 70

func fetchNewDeposits(
	ctx context.Context,
	liquidity *bindings.Liquidity,
	startBlock uint64,
) (_ []*bindings.LiquidityDeposited, _ uint64, _ map[uint32]bool, _ error) {
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

func fetchAMLScore(sender string, contractAddress string) uint32 { // nolint:gocritic
	const int50Key = 50
	// TODO: Implement a real AML score fetching function
	return int50Key
}

func rejectDeposits(
	ctx context.Context,
	cfg *configs.Config,
	client *ethclient.Client,
	liquidity *bindings.Liquidity,
	maxLastSeenDepositIndex uint64,
	rejectedIndices []uint64,
) (*types.Receipt, error) {
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

func DepositAnalyzer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) {
	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var client *ethclient.Client
	client, err = newClient(link)
	if err != nil {
		log.Fatalf(err.Error())
	}

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		log.Fatalf("Failed to instantiate a Liquidity contract: %v", err.Error())
	}

	// example -- start
	const (
		msgTestF = "Failed to fetch token by tokenIndex with DBApp: %v"
		msgTestC = "Failed to create token by tokenIndex with DBApp: %v"
		tokenIdx = "adasdasdsa"
	)
	var token *mDBApp.Token
	token, err = db.TokenByIndex(tokenIdx)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		panic(fmt.Sprintf(msgTestF, err.Error()))
	}
	log.Printf("----------- 1 %+v\n", token)

	const int11111Key = 11111
	var tokenID uint256.Int
	_ = tokenID.SetFromBig(new(big.Int).SetInt64(int64(int11111Key)))
	var vvv *mDBApp.Token
	vvv, err = db.CreateToken(tokenIdx, "", &tokenID)
	if err != nil {
		panic(fmt.Sprintf(msgTestC, err.Error()))
	}
	log.Printf("---------- 2 %+v\n", vvv)

	token, err = db.TokenByIndex(tokenIdx)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		panic(fmt.Sprintf(msgTestF, err.Error()))
	}
	if errors.Is(err, errorsDB.ErrNotFound) {
		panic(fmt.Sprintf(msgTestF, err.Error()))
	}
	log.Printf("----------- 3 %+v\n", token)
	// example -- finish

	lastSeenDepositIndex, err := liquidity.GetLastSeenDepositIndex(&bind.CallOpts{
		Pending: false,
		Context: ctx,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to get last seen deposit index: %v", err.Error()))
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
