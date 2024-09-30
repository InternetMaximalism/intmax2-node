package deposit_synchronizer

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/utils"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var ErrStatCurrentFileFail = errors.New("stat current file fail")

type depositSynchronizer struct {
	ctx          context.Context
	cfg          *configs.Config
	log          logger.Logger
	dbApp        SQLDriverApp
	sb           ServiceBlockchain
	scrollClient *ethclient.Client
	rollup       *bindings.Rollup
}

func New(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	dbApp SQLDriverApp,
	sb ServiceBlockchain,
) (DepositSynchronizer, error) {
	scrollLink, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Scroll network chain link: %w", err)
	}

	scrollClient, err := utils.NewClient(scrollLink)
	if err != nil {
		return nil, fmt.Errorf("failed to create new scrollClient: %w", err)
	}
	defer scrollClient.Close()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.Blockchain.RollupContractAddress), scrollClient)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	return &depositSynchronizer{
		ctx:          ctx,
		cfg:          cfg,
		log:          log,
		dbApp:        dbApp,
		sb:           sb,
		scrollClient: scrollClient,
		rollup:       rollup,
	}, nil
}

func (w *depositSynchronizer) Init(
	ctx context.Context,
) (err error) {
	const (
		hName = "DepositSynchronizer func:Init"
	)

	spanCtx, span := open_telemetry.Tracer().Start(w.ctx, hName)
	defer span.End()

	err = w.sb.SetupScrollNetworkChainID(w.ctx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrSetupScrollNetworkChainIDFail, err)
	}

	return nil
}

func (w *depositSynchronizer) Start(
	tickerEventWatcher *time.Ticker,
) (err error) {
	const (
		hName = "DepositSynchronizer func:Start"
	)

	spanCtx, span := open_telemetry.Tracer().Start(w.ctx, hName)
	defer span.End()

	err = w.Init(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrInitFail, err)
	}

	var link string
	link, err = w.sb.ScrollNetworkChainLinkEvmJSONRPC(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	rollupCfg := intMaxTypes.NewRollupContractConfigFromEnv(w.cfg, link)
	if w.cfg.Blockchain.BlockSynchronizerHex != "" {
		rollupCfg.EthereumPrivateKeyHex = w.cfg.Blockchain.BlockSynchronizerHex
	}

	for {
		select {
		case <-w.ctx.Done():
			return nil
		case <-tickerEventWatcher.C:
			shouldProcess := func() (bool, error) {
				latestBlock, err := intMaxTypes.FetchLatestIntMaxBlock(rollupCfg, w.ctx)
				if err != nil {
					if err.Error() != "no posted blocks found" {
						return false, err
					}

					return true, nil
				}
				latestDepositTreeRoot, err := intMaxTypes.FetchDepositRoot(rollupCfg, w.ctx)
				if err != nil {
					return false, err
				}

				if latestBlock.DepositTreeRoot == latestDepositTreeRoot {
					return false, nil
				}

				// TODO: Check that no new blocks have been created for 15 minutes
				// after the deposit tree root has been communicated to the Rollup contract.

				return true, nil
			}

			var ok bool
			ok, err = shouldProcess()
			if err != nil {
				return err
			}
			if !ok {
				fmt.Printf("No new deposits\n")
				continue
			}

			// Generate a new empty block to reflect new deposits.
			// TODO: If there is a block already in the process of being created, there is no need to post this block.
			blockContent := block_post_service.MakeEmptyBlock()

			_, err = intMaxTypes.MakePostRegistrationBlockInput(
				blockContent,
			)
			if err != nil {
				return err
			}

			_, err = intMaxTypes.PostRegistrationBlock(rollupCfg, w.ctx, w.log, w.scrollClient, blockContent)
			if err != nil {
				return err
			}
		}
	}
}

// func FetchLastWitnessGeneratedBlockNumber(rollup *bindings.Rollup, ctx context.Context) (uint32, error) {
// 	opts := bind.CallOpts{
// 		Pending: false,
// 		Context: ctx,
// 	}
// 	latestBlockNumber, err := rollup.GetLatestBlockNumber(&opts)

// 	return latestBlockNumber, err
// }

// const int32Key = 32

// func FetchLatestIntMaxBlock(rollup *bindings.Rollup, ctx context.Context) (*bindings.RollupBlockPosted, error) {
// 	latestBlockNumber, err := FetchLastWitnessGeneratedBlockNumber(rollup, ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch latest block number: %w", err)
// 	}
// 	if latestBlockNumber == 0 {
// 		defaultDepositTreeRoot := [int32Key]byte{}
// 		var decodedDefaultDepositTreeRoot []byte
// 		decodedDefaultDepositTreeRoot, err = hexutil.Decode("0xb6155ab566bbd2e341525fd88c43b4d69572bf4afe7df45cd74d6901a172e41c")
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to decode default deposit tree root: %w", err)
// 		}

// 		copy(defaultDepositTreeRoot[:], decodedDefaultDepositTreeRoot)
// 		return &bindings.RollupBlockPosted{
// 			PrevBlockHash:   [int32Key]byte{},
// 			BlockBuilder:    common.Address{},
// 			BlockNumber:     big.NewInt(0),
// 			DepositTreeRoot: defaultDepositTreeRoot,
// 			SignatureHash:   [int32Key]byte{},
// 		}, nil
// 	}

// 	latestPrevBlockHash, err := FetchIntMaxBlock(cfg, ctx, latestBlockNumber-1)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch latest block: %w", err)
// 	}

// 	blocks, _, err := FetchPostedBlocks(cfg, ctx, cfg.RollupContractDeployedBlockNumber, [][int32Key]byte{latestPrevBlockHash}, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch posted blocks: %w", err)
// 	}
// 	if len(blocks) == 0 {
// 		return nil, errors.New("no posted blocks found")
// 	}

// 	return blocks[0], nil
// }

// func FetchDepositRoot(ctx context.Context, rollup *bindings.Rollup) (*bindings.RollupBlockPosted, error) {
// 	opts := bind.CallOpts{
// 		Pending: false,
// 		Context: ctx,
// 	}
// 	latestDepositTreeRoot, err := rollup.GetLatestDepositTreeRoot(&opts)

// 	return latestDepositTreeRoot, err
// }
