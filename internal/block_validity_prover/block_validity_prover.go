package block_validity_prover

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/logger"
	"intmax2-node/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	relayMessageMethod    = "relayMessage"    // "8ef1332e"
	processDepositsMethod = "processDeposits" // "f03efa37"
	eventBlockRange       = 100000
)

type blockValidityProver struct {
	ctx          context.Context
	cfg          *configs.Config
	log          logger.Logger
	ethClient    *ethclient.Client
	scrollClient *ethclient.Client
	liquidity    *bindings.Liquidity
	rollup       *bindings.Rollup
	blockBuilder BlockBuilderStorage
}

func NewBlockValidityProver(ctx context.Context, cfg *configs.Config, log logger.Logger, sb ServiceBlockchain, db SQLDriverApp) (*blockValidityProver, error) {
	ethClient, err := utils.NewClient(cfg.Blockchain.EthereumNetworkRpcUrl)
	if err != nil {
		return nil, errors.Join(ErrNewEthereumClientFail, err)
	}
	defer ethClient.Close()

	scrollLink, err := sb.ScrollNetworkChainLinkEvmJSONRPC(ctx)
	if err != nil {
		return nil, errors.Join(ErrScrollNetwrokChainLink, err)
	}

	scrollClient, err := utils.NewClient(scrollLink)
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

	blockBuilder := NewMockBlockBuilder(cfg, db)

	return &blockValidityProver{
		ctx:          ctx,
		cfg:          cfg,
		log:          log,
		ethClient:    ethClient,
		scrollClient: scrollClient,
		liquidity:    liquidity,
		rollup:       rollup,
		blockBuilder: blockBuilder,
	}, nil
}

func (d *blockValidityProver) FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error) {
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

func (d *blockValidityProver) FetchLastDepositIndex() (uint32, error) {
	return d.blockBuilder.FetchLastDepositIndex()
}
