//nolint:gocritic
package deposit_service

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"

	"github.com/jackc/pgx/v5"

	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"intmax2-node/pkg/utils"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const amlRejectionThreshold = 70
const noDepositEventsFoundError = "No deposit events found"

type DepositAnalyzerService struct {
	ctx       context.Context
	cfg       *configs.Config
	log       logger.Logger
	client    *ethclient.Client
	liquidity *bindings.Liquidity
}

func NewDepositAnalyzerService(ctx context.Context, cfg *configs.Config, log logger.Logger, sc ServiceBlockchain) (*DepositAnalyzerService, error) {
	return newDepositAnalyzerService(ctx, cfg, log, sc)
}

func newDepositAnalyzerService(ctx context.Context, cfg *configs.Config, log logger.Logger, _ ServiceBlockchain) (*DepositAnalyzerService, error) {
	// link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	// if err != nil {
	//   return nil, fmt.Errorf("failed to get Ethereum network chain link: %w", err)
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

	event, err := db.EventBlockNumberByEventName(mDBApp.DepositsAnalyzedEvent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "not found" {
			event = &mDBApp.EventBlockNumber{
				EventName:                mDBApp.DepositsAnalyzedEvent,
				LastProcessedBlockNumber: 0,
			}
		} else {
			panic(fmt.Sprintf("Error fetching event block number: %v", err.Error()))
		}
	} else if event == nil {
		event = &mDBApp.EventBlockNumber{
			EventName:                mDBApp.DepositsAnalyzedEvent,
			LastProcessedBlockNumber: 0,
		}
	}

	lastEventInfo, err := depositAnalyzerService.fetchLastDepositAnalyzedEvent(event.LastProcessedBlockNumber)
	if err != nil {
		panic(fmt.Sprintf("Failed to get last deposit analyzed block number: %v", err.Error()))
	}
	if lastEventInfo == nil || lastEventInfo.BlockNumber == nil {
		panic("Last event info or block number is nil")
	}

	events, maxDepositIndex, tokenIndexMap, err := depositAnalyzerService.fetchNewDeposits(*lastEventInfo.BlockNumber)
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
		return
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
			rejectDepositIndices = append(rejectDepositIndices, new(big.Int).SetUint64(uint64(event.TokenIndex)))
		}
	}

	receipt, err := depositAnalyzerService.analyzeDeposits(maxDepositIndex, rejectDepositIndices)
	if err != nil {
		panic(fmt.Sprintf("Failed to analyze deposits: %v", err.Error()))
	}

	if receipt == nil {
		panic("Received nil receipt for transaction")
	}

	switch receipt.Status {
	case types.ReceiptStatusSuccessful:
		log.Infof("Successfully deposit analyzed %d deposits, %d rejections. Transaction Hash: %v", len(events), len(rejectDepositIndices), receipt.TxHash.Hex())
	case types.ReceiptStatusFailed:
		panic(fmt.Sprintf("Transaction failed: deposit analyzed unsuccessful. Transaction Hash: %v", receipt.TxHash.Hex()))
	default:
		panic(fmt.Sprintf("Unexpected transaction status: %d. Transaction Hash: %v", receipt.Status, receipt.TxHash.Hex()))
	}

	_, err = db.UpsertEventBlockNumber(mDBApp.DepositsAnalyzedEvent, *lastEventInfo.BlockNumber)
	if err != nil {
		panic(fmt.Sprintf("Error updating event block number: %v", err.Error()))
	}
}

func (d *DepositAnalyzerService) fetchLastDepositAnalyzedEvent(startBlockNumber uint64) (*DepositEventInfo, error) {
	iterator, err := d.liquidity.FilterDepositsAnalyzed(&bind.FilterOpts{
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

func (d *DepositAnalyzerService) fetchNewDeposits(startBlock uint64) (_ []*bindings.LiquidityDeposited, _ *big.Int, _ map[uint32]bool, _ error) {
	nextBlock := startBlock + 1
	iterator, err := d.liquidity.FilterDeposited(&bind.FilterOpts{
		Start:   nextBlock,
		End:     nil,
		Context: d.ctx,
	}, []*big.Int{}, []common.Address{})
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
	eventInfo, err := fetchDepositEvent(d.liquidity, lastBlockNumber, depositIds)
	if err != nil {
		if err.Error() == noDepositEventsFoundError {
			fmt.Println("No deposit events found, skipping process")
			return false, nil
		}
		return false, fmt.Errorf("failed to get last block number: %w", err)
	}
	fmt.Println("eventInfo ", eventInfo)

	if *eventInfo.BlockNumber == 0 {
		return false, nil
	}

	isExceeded, err := isBlockTimeExceeded(d.client, *eventInfo.BlockNumber, int(d.cfg.Blockchain.DepositAnalyzerMinutesThreshold))
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

func (d *DepositAnalyzerService) analyzeDeposits(upToDepositId *big.Int, rejectDepositIndices []*big.Int) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(d.cfg.Blockchain.DepositAnalyzerPrivateKeyHex, d.cfg.Blockchain.EthereumNetworkChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	tx, err := d.liquidity.AnalyzeDeposits(transactOpts, upToDepositId, rejectDepositIndices)
	if err != nil {
		return nil, fmt.Errorf("failed to send AnalyzeDeposits transaction: %w", err)
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
