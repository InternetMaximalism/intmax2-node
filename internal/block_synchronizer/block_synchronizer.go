package block_synchronizer

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	scrollNetworkRpcUrl   = "https://sepolia-rpc.scroll.io"
	senderPublicKeysIndex = 5
	numOfSenders          = intMaxTypes.NumOfSenders

	int1Key   = 1
	int32Key  = 32
	minus1Key = -1
)

type blockSynchronizer struct {
	ctx          context.Context
	cfg          *configs.Config
	log          logger.Logger
	ethClient    *ethclient.Client
	scrollClient *ethclient.Client
	liquidity    *bindings.Liquidity
	rollup       *bindings.Rollup
}

func NewBlockSynchronizer(ctx context.Context, cfg *configs.Config, log logger.Logger) (BlockSynchronizer, error) {
	ethClient, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, errors.Join(ErrNewEthereumClientFail, err)
	}
	defer ethClient.Close()

	var scrollClient *ethclient.Client
	scrollClient, err = utils.NewClient(scrollNetworkRpcUrl)
	if err != nil {
		return nil, errors.Join(ErrNewScrollClientFail, err)
	}
	defer scrollClient.Close()

	var liquidity *bindings.Liquidity
	liquidity, err = bindings.NewLiquidity(
		common.HexToAddress(cfg.Blockchain.LiquidityContractAddress),
		ethClient,
	)
	if err != nil {
		return nil, errors.Join(ErrInstantiateLiquidityContractFail, err)
	}

	var rollup *bindings.Rollup
	rollup, err = bindings.NewRollup(
		common.HexToAddress(cfg.Blockchain.RollupContractAddress),
		scrollClient,
	)
	if err != nil {
		return nil, errors.Join(ErrInstantiateRollupContractFail, err)
	}

	return &blockSynchronizer{
		ctx:          ctx,
		cfg:          cfg,
		log:          log,
		ethClient:    ethClient,
		scrollClient: scrollClient,
		liquidity:    liquidity,
		rollup:       rollup,
	}, nil
}

func (d *blockSynchronizer) RollupContractDeployedBlockNumber() uint64 {
	return d.cfg.Blockchain.RollupContractDeployedBlockNumber
}

func (d *blockSynchronizer) FetchLatestBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := d.scrollClient.BlockNumber(ctx)
	if err != nil {
		return 0, errors.Join(ErrFetchLatestBlockNumberFail, err)
	}

	return blockNumber, nil
}

func (d *blockSynchronizer) FetchNewPostedBlocks(startBlock uint64, endBlock *uint64) ([]*bindings.RollupBlockPosted, *big.Int, error) {
	nextBlock := startBlock + int1Key
	if endBlock != nil && nextBlock > *endBlock {
		return nil, nil, nil
	}

	iterator, err := d.rollup.FilterBlockPosted(&bind.FilterOpts{
		Start:   nextBlock,
		End:     endBlock,
		Context: d.ctx,
	}, [][int32Key]byte{}, []common.Address{})
	if err != nil {
		return nil, nil, errors.Join(ErrFilterLogsFail, err)
	}

	defer func() {
		_ = iterator.Close()
	}()

	var events []*bindings.RollupBlockPosted
	maxBlockNumber := new(big.Int).SetUint64(startBlock)

	for iterator.Next() {
		event := iterator.Event
		events = append(events, event)
		currBN := new(big.Int).SetUint64(event.Raw.BlockNumber)
		if maxBlockNumber.Cmp(currBN) == minus1Key {
			maxBlockNumber.Set(currBN)
		}
	}

	if err = iterator.Error(); err != nil {
		return nil, nil, errors.Join(ErrEncounteredWhileIterating, err)
	}

	return events, maxBlockNumber, nil
}

func (d *blockSynchronizer) FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error) {
	tx, isPending, err := d.scrollClient.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, errors.Join(ErrTransactionByHashNotFound, err)
	}

	if isPending {
		return nil, ErrTransactionIsStillPending
	}

	calldata := tx.Data()

	return calldata, nil
}

// func (d *blockPostService) FetchNewDeposits(startBlock uint64) ([]*bindings.LiquidityDeposited, *big.Int, map[uint32]bool, error) {
// 	nextBlock := startBlock + 1
// 	iterator, err := d.liquidity.FilterDeposited(&bind.FilterOpts{
// 		Start:   nextBlock,
// 		End:     nil,
// 		Context: d.ctx,
// 	}, []*big.Int{}, []common.Address{})
// 	if err != nil {
// 		return nil, nil, nil, errors.Join(ErrFilterLogsFail, err)
// 	}

// 	defer iterator.Close()

// 	var events []*bindings.LiquidityDeposited
// 	maxDepositIndex := new(big.Int)
// 	tokenIndexMap := make(map[uint32]bool)

// 	for iterator.Next() {
// 		event := iterator.Event
// 		events = append(events, event)
// 		tokenIndexMap[event.TokenIndex] = true
// 		if event.DepositId.Cmp(maxDepositIndex) > 0 {
// 			maxDepositIndex.Set(event.DepositId)
// 		}
// 	}

// 	if err = iterator.Error(); err != nil {
// 		return nil, nil, nil, errors.Join(ErrEncounteredWhileIterating, err)
// 	}

// 	return events, maxDepositIndex, tokenIndexMap, nil
// }
