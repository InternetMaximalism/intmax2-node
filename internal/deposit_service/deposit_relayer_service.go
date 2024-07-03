package deposit_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"math/big"
	"time"

	"intmax2-node/pkg/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const depositThreshold uint64 = 128
const duration = 1 * time.Hour

type DepositIndices struct {
	LastSeenDepositIndex      *uint64
	LastProcessedDepositIndex *uint64
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

func fetchDepositIndices(ctx context.Context, liquidity *bindings.Liquidity) (DepositIndices, error) {
	type result struct {
		index uint64
		err   error
	}

	opts := &bind.CallOpts{
		Pending: false,
		Context: ctx,
	}

	lastSeenDepositIndexCh := make(chan result)
	lastProcessedDepositIndexCh := make(chan result)

	go func() {
		index, err := liquidity.GetLastSeenDepositIndex(opts)
		lastSeenDepositIndexCh <- result{index, err}
	}()

	go func() {
		index, err := liquidity.GetLastProcessedDepositIndex(opts)
		lastProcessedDepositIndexCh <- result{index, err}
	}()

	var di DepositIndices
	for {
		if di.LastSeenDepositIndex != nil && di.LastProcessedDepositIndex != nil {
			return di, nil
		}
		select {
		case lastSeenDepositResult := <-lastSeenDepositIndexCh:
			if lastSeenDepositResult.err != nil {
				return DepositIndices{}, lastSeenDepositResult.err
			}
			di.LastSeenDepositIndex = &lastSeenDepositResult.index
		case lastProcessedDepositResult := <-lastProcessedDepositIndexCh:
			if lastProcessedDepositResult.err != nil {
				return DepositIndices{}, lastProcessedDepositResult.err
			}
			di.LastProcessedDepositIndex = &lastProcessedDepositResult.index
		}
	}
}

func submitDepositRoot(ctx context.Context, cfg *configs.Config, client *ethclient.Client, liquidity *bindings.Liquidity, maxLastSeenDepositIndex uint64) (*types.Receipt, error) {
	transactOpts, err := utils.CreateTransactor(cfg)
	if err != nil {
		return nil, err
	}

	tx, err := liquidity.SubmitDepositRoot(transactOpts, maxLastSeenDepositIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to send RejectDeposits transaction: %w", err)
	}

	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return receipt, nil
}

func shouldProcessDeposits(client *ethclient.Client, liquidity *bindings.Liquidity, lastSeenDepositIndex, lastProcessedDepositIndex uint64) (bool, error) {
	unprocessedDepositCount := lastSeenDepositIndex - lastProcessedDepositIndex
	if unprocessedDepositCount <= 0 {
		return false, nil
	}

	if unprocessedDepositCount >= depositThreshold {
		fmt.Println("Deposit threshold is reached")
		return true, nil
	}

	nextDepositIndex := lastProcessedDepositIndex + 1
	lastProcessedBlockNumber, err := utils.FetchBlockNumberByDepositIndex(liquidity, nextDepositIndex)
	if err != nil {
		return false, fmt.Errorf("failed to get last block number: %w", err)
	}

	// If there is no last processed block number, it means that there is no deposit
	if lastProcessedBlockNumber == 0 {
		return false, nil
	}

	isExceeded, err := isBlockTimeExceeded(client, lastProcessedBlockNumber)
	if err != nil {
		return false, fmt.Errorf("error occurred while checking time difference: %w", err)
	}

	if !isExceeded {
		return false, nil
	}

	fmt.Println("Block time difference exceeded the specified duration")
	return true, nil
}

func DepositRelayer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) {
	link, err := sb.EthereumNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var client *ethclient.Client
	client, err = utils.NewClient(link)
	if err != nil {
		log.Fatalf(err.Error())
	}

	liquidity, err := bindings.NewLiquidity(common.HexToAddress(cfg.Blockchain.LiquidityContractAddress), client)
	if err != nil {
		log.Fatalf("Failed to instantiate a Liquidity contract: %v", err.Error())
	}

	indices, err := fetchDepositIndices(ctx, liquidity)
	if err != nil {
		log.Fatalf("Failed to fetch deposit indices: %v", err.Error())
	}

	shouldSubmit, err := shouldProcessDeposits(client, liquidity, *indices.LastSeenDepositIndex, *indices.LastProcessedDepositIndex)
	if err != nil {
		log.Fatalf("Error in threshold and time diff check: %v", err)
		return
	}

	if !shouldSubmit {
		return
	}

	receipt, err := submitDepositRoot(ctx, cfg, client, liquidity, *indices.LastSeenDepositIndex)
	if err != nil {
		log.Fatalf("Failed to submit deposit root: %v", err.Error())
	}

	if receipt == nil {
		return
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Infof("Successfully submitted deposit root")
	} else {
		log.Infof("Failed to submit deposit root")
	}

	log.Infof("Tx Hash: %v", receipt.TxHash.String())
}
