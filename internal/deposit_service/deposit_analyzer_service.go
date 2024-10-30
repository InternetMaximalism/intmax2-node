//nolint:gocritic
package deposit_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"time"

	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"intmax2-node/pkg/utils"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const fixedDepositValueInWei = 1e17 // 0.1 ETH in Wei
const amlRejectionThreshold = 70
const noDepositEventsFoundError = "No deposit events found"

type DepositAnalyzerService struct {
	ctx       context.Context
	cfg       *configs.Config
	log       logger.Logger
	client    *ethclient.Client
	liquidity *bindings.Liquidity
}

type DepositEventInfo struct {
	LastDepositId *uint64
	BlockNumber   *uint64
}

func NewDepositAnalyzerService(ctx context.Context, cfg *configs.Config, log logger.Logger, sc ServiceBlockchain) (*DepositAnalyzerService, error) {
	return newDepositAnalyzerService(ctx, cfg, log, sc)
}

func newDepositAnalyzerService(ctx context.Context, cfg *configs.Config, log logger.Logger, sb ServiceBlockchain) (*DepositAnalyzerService, error) {
	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
	}

	client, err := utils.NewClient(link)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	return &DepositAnalyzerService{
		ctx:       ctx,
		cfg:       cfg,
		log:       log,
		client:    client,
		liquidity: liquidity,
	}, nil
}

// TODO: TxManager Class that stops processing if there are any pending transactions.
func DepositAnalyzer(ctx context.Context, cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb ServiceBlockchain) {
	depositAnalyzerService, err := newDepositAnalyzerService(ctx, cfg, log, sb)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DepositAnalyzerService: %v", err.Error()))
	}

	_ = db.Exec(ctx, nil, func(d interface{}, _ interface{}) (err error) {
		q := d.(SQLDriverApp)

		event, err := q.EventBlockNumberByEventName(mDBApp.DepositsAndAnalyzedRelayedEvent)
		if err != nil {
			if errors.Is(err, errorsDB.ErrNotFound) {
				event = &mDBApp.EventBlockNumber{
					EventName:                mDBApp.DepositsAndAnalyzedRelayedEvent,
					LastProcessedBlockNumber: cfg.Blockchain.LiquidityContractDeployedBlockNumber,
				}
			} else {
				panic(fmt.Sprintf("Error fetching event block number: %v", err.Error()))
			}
		} else if event == nil {
			event = &mDBApp.EventBlockNumber{
				EventName:                mDBApp.DepositsAndAnalyzedRelayedEvent,
				LastProcessedBlockNumber: cfg.Blockchain.LiquidityContractDeployedBlockNumber,
			}
		}

		lastEventInfo, err := depositAnalyzerService.fetchLastDepositsAndAnalyzedReleyedEvent(event.LastProcessedBlockNumber)
		if err != nil {
			panic(fmt.Sprintf("Failed to get last deposit analyzed block number: %v", err.Error()))
		}

		if lastEventInfo == nil || lastEventInfo.BlockNumber == nil {
			panic("Last event info or block number is nil")
		}

		if *lastEventInfo.BlockNumber == uint64(0) {
			lastEventInfo.BlockNumber = &event.LastProcessedBlockNumber
		}

		_, err = q.UpsertEventBlockNumber(mDBApp.DepositsAndAnalyzedRelayedEvent, *lastEventInfo.BlockNumber)
		if err != nil {
			panic(fmt.Sprintf("Error updating event block number: %v", err.Error()))
		}

		var (
			events          []*bindings.LiquidityDeposited
			maxDepositIndex *big.Int
			tokenIndexMap   map[uint32]bool
		)

		events, maxDepositIndex, tokenIndexMap, err =
			depositAnalyzerService.fetchNewDeposits(*lastEventInfo.BlockNumber)
		if err != nil {
			panic(fmt.Sprintf("Failed to fetch new deposits: %v", err.Error()))
		}

		shouldSubmit, err := depositAnalyzerService.shouldProcessDepositAnalyzer(
			events,
			*lastEventInfo.BlockNumber,
		)
		if err != nil {
			panic(fmt.Sprintf("Error in threshold and time diff check: %v", err.Error()))
		}

		if !shouldSubmit {
			log.Infof(
				"Deposit analyzer will not be processed at this time. Unprocessed deposit count: %d, Last deposit block number: %d",
				len(events),
				*lastEventInfo.BlockNumber,
			)
			return nil
		}

		tokenInfoMap, err := depositAnalyzerService.getTokenInfoMap(tokenIndexMap)
		if err != nil {
			panic(fmt.Sprintf("Failed to get token info map: %v", err.Error()))
		}

		var rejectDepositIndices []*big.Int
		for _, event := range events {
			contractAddress := tokenInfoMap[event.TokenIndex]
			score := fetchAMLScore(event.Sender.Hex(), contractAddress.Hex())
			if score > amlRejectionThreshold {
				rejectDepositIndices = append(rejectDepositIndices, new(big.Int).SetUint64(event.DepositId.Uint64()))
			}
		}

		lastRelayedDepositId, err := depositAnalyzerService.getLastRelayedDepositId()
		if err != nil {
			panic(fmt.Sprintf("Failed to get last relayed deposit id: %v", err.Error()))
		}

		pendingDepositsCount := maxDepositIndex.Uint64() - lastRelayedDepositId - uint64(len(rejectDepositIndices))
		receipt, err := depositAnalyzerService.analyzeAndRelayDeposits(maxDepositIndex, rejectDepositIndices, pendingDepositsCount)
		if err != nil {
			panic(fmt.Sprintf("Failed to analyze and relay deposits: %v", err.Error()))
		}

		if receipt == nil {
			panic("Received nil receipt for transaction")
		}

		switch receipt.Status {
		case types.ReceiptStatusSuccessful:
			log.Infof("Successfully analyzed and relayed deposits %d deposits, %d rejections. Transaction Hash: %v", len(events), len(rejectDepositIndices), receipt.TxHash.Hex())
		case types.ReceiptStatusFailed:
			panic(fmt.Sprintf("Transaction failed: analyzed and relayed deposits unsuccessful. Transaction Hash: %v", receipt.TxHash.Hex()))
		default:
			panic(fmt.Sprintf("Unexpected transaction status: %d. Transaction Hash: %v", receipt.Status, receipt.TxHash.Hex()))
		}

		return nil
	})
}

func (d *DepositAnalyzerService) fetchLastDepositsAndAnalyzedReleyedEvent(startBlockNumber uint64) (*DepositEventInfo, error) {
	nextBlock := startBlockNumber + 1
	iterator, err := d.liquidity.FilterDepositsAnalyzedAndRelayed(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: d.ctx,
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

		currentId := iterator.Event.UpToDepositId.Uint64()
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

func (d *DepositAnalyzerService) fetchNewDeposits(
	startBlock uint64,
) ([]*bindings.LiquidityDeposited, *big.Int, map[uint32]bool, error) {
	nextBlock := startBlock + 1
	iterator, err := d.liquidity.FilterDeposited(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: d.ctx,
	}, []*big.Int{}, []common.Address{}, [][32]byte{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to filter logs: %w", err)
	}

	defer iterator.Close()

	var events []*bindings.LiquidityDeposited
	maxDepositIndex := new(big.Int)
	tokenIndexMap := make(map[uint32]bool)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
		tokenIndexMap[event.TokenIndex] = true
		if event.DepositId.Cmp(maxDepositIndex) > 0 {
			maxDepositIndex.Set(event.DepositId)
		}
	}

	if err = iterator.Error(); err != nil {
		return nil, nil, nil, fmt.Errorf("error encountered while iterating: %w", err)
	}

	return events, maxDepositIndex, tokenIndexMap, nil
}

func (d *DepositAnalyzerService) shouldProcessDepositAnalyzer(events []*bindings.LiquidityDeposited, lastBlockNumber uint64) (bool, error) {
	eventCount := len(events)
	if eventCount <= 0 {
		return false, nil
	}
	if eventCount >= int(d.cfg.Blockchain.DepositAnalyzerThreshold) {
		d.log.Infof("Deposit analyzer threshold is reached: %d", eventCount)
		return true, nil
	}

	depositIds := []*big.Int{}
	eventInfo, err := fetchDepositEvent(d.ctx, d.liquidity, lastBlockNumber, depositIds)
	if err != nil {
		if err.Error() == noDepositEventsFoundError {
			fmt.Println("No deposit events found, skipping process")
			return false, nil
		}
		return false, fmt.Errorf("failed to get last block number: %w", err)
	}

	if *eventInfo.BlockNumber == 0 {
		return false, nil
	}

	isExceeded, err := isBlockTimeExceeded(d.ctx, d.client, *eventInfo.BlockNumber, int(d.cfg.Blockchain.DepositAnalyzerMinutesThreshold))
	if err != nil {
		return false, fmt.Errorf("error occurred while checking time difference: %w", err)
	}

	if !isExceeded {
		return false, nil
	}

	fmt.Println("Block time difference exceeded the specified duration")
	return true, nil
}

func (d *DepositAnalyzerService) getTokenInfoMap(tokenIndexMap map[uint32]bool) (map[uint32]common.Address, error) {
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
			tokenInfo, err := d.liquidity.GetTokenInfo(&bind.CallOpts{
				Pending: false,
				Context: d.ctx,
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

func (d *DepositAnalyzerService) analyzeAndRelayDeposits(upToDepositId *big.Int, rejectDepositIndices []*big.Int, numDepositsToRelay uint64) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(d.cfg.Blockchain.DepositAnalyzerPrivateKeyHex, d.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	transactOpts.Value = big.NewInt(fixedDepositValueInWei)
	gasLimit := new(big.Int).SetUint64(calculateAnalyzeAndRelayGasLimit(numDepositsToRelay))

	err = utils.LogTransactionDebugInfo(
		d.log,
		d.cfg.Blockchain.DepositAnalyzerPrivateKeyHex,
		d.cfg.Blockchain.LiquidityContractAddress,
		upToDepositId,
		rejectDepositIndices,
		gasLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to log transaction debug info: %w", err)
	}

	tx, err := d.liquidity.AnalyzeAndRelayDeposits(transactOpts, upToDepositId, rejectDepositIndices, gasLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to send AnalyzeAndRelayDeposits transaction: %w", err)
	}

	receipt, err := bind.WaitMined(d.ctx, d.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func fetchAMLScore(sender string, contractAddress string) uint32 { // nolint:gocritic
	const int50Key = 50
	// TODO: Implement a real AML score fetching function
	return int50Key
}

func calculateAnalyzeAndRelayGasLimit(numDepositsToRelay uint64) uint64 {
	const (
		baseGas       = uint64(220000)
		perDepositGas = uint64(20000)
		bufferGas     = uint64(100000)
	)
	return baseGas + (perDepositGas * numDepositsToRelay) + bufferGas
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

func fetchDepositEvent(ctx context.Context, liquidity *bindings.Liquidity, startBlockNumber uint64, depositIds []*big.Int) (*DepositEventInfo, error) {
	nextBlock := startBlockNumber + 1
	iterator, err := liquidity.FilterDeposited(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: ctx,
	}, depositIds, []common.Address{}, [][32]byte{})
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

func (d *DepositAnalyzerService) getLastRelayedDepositId() (uint64, error) {
	lastRelayedDepositId, err := d.liquidity.GetLastRelayedDepositId(&bind.CallOpts{
		Pending: false,
		Context: d.ctx,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get last relayed deposit id: %w", err)
	}
	if !lastRelayedDepositId.IsUint64() {
		return 0, fmt.Errorf("last relayed deposit id exceeds uint64 range")
	}
	return lastRelayedDepositId.Uint64(), nil
}
