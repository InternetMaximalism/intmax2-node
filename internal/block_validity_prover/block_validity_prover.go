package block_validity_prover

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
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
	blockBuilder *mockBlockBuilder
}

type BlockValidityProverMemory = blockValidityProver

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

func (d *blockValidityProver) LastSeenBlockPostedEventBlockNumber() (uint64, error) {
	return d.blockBuilder.LastSeenBlockPostedEventBlockNumber()
}

func (d *blockValidityProver) SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error {
	return d.blockBuilder.SetLastSeenBlockPostedEventBlockNumber(blockNumber)
}

func (d *blockValidityProver) LatestIntMaxBlockNumber() uint32 {
	return d.blockBuilder.LatestIntMaxBlockNumber()
}

func (d *blockValidityProver) GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *DepositLeafWithId, depositIndex *uint32, err error) {
	return d.blockBuilder.GetDepositLeafAndIndexByHash(depositHash)
}

func (d *blockValidityProver) BlockNumberByDepositIndex(depositIndex uint32) (uint32, error) {
	// TODO: implement this method
	return d.blockBuilder.BlockNumberByDepositIndex(depositIndex)
}

func (d *blockValidityProver) IsSynchronizedDepositIndex(depositIndex uint32) (bool, error) {
	return d.blockBuilder.IsSynchronizedDepositIndex(depositIndex)
}

func (d *blockValidityProver) FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error) {
	return d.blockBuilder.FetchUpdateWitness(publicKey, currentBlockNumber, targetBlockNumber, isPrevAccountTree)
}

func (d *blockValidityProver) BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.MerkleProof, error) {
	return d.blockBuilder.BlockTreeProof(rootBlockNumber, leafBlockNumber)
}

func (d *blockValidityProver) PostBlock(
	isRegistrationBlock bool,
	txs []*MockTxRequest,
) (*ValidityWitness, error) {
	return d.blockBuilder.PostBlock(isRegistrationBlock, txs)
}

func (d *blockValidityProver) BlockContentByTxRoot(txRoot string) (*block_post_service.PostedBlock, error) {
	blockContent, err := d.blockBuilder.BlockContentByTxRoot(txRoot)
	if err != nil {
		var ErrBlockContentByTxRoot = errors.New("failed to get block content by tx root")
		return nil, errors.Join(ErrBlockContentByTxRoot, err)
	}

	if blockContent.PrevBlockHash == "" {
		var ErrPrevBlockHash = errors.New("prev block hash is empty")
		return nil, errors.Join(ErrPrevBlockHash, err)
	}
	prevBlockHash := common.HexToHash("0x" + blockContent.PrevBlockHash)

	if blockContent.DepositRoot == "" {
		var ErrDepositRoot = errors.New("deposit root is empty")
		return nil, errors.Join(ErrDepositRoot, err)
	}
	depositRoot := common.HexToHash("0x" + blockContent.DepositRoot)

	if blockContent.SignatureHash == "" {
		var ErrSignatureHash = errors.New("signature hash is empty")
		return nil, errors.Join(ErrSignatureHash, err)
	}
	signatureHash := common.HexToHash("0x" + blockContent.SignatureHash)

	return block_post_service.NewPostedBlock(
		prevBlockHash,
		depositRoot,
		blockContent.BlockNumber,
		signatureHash,
	), nil
}

func (d *blockValidityProver) DepositTreeProof(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	lastValidityWitness, err := d.blockBuilder.LastValidityWitness()
	if err != nil {
		return nil, common.Hash{}, errors.New("last validity witness not found")
	}
	blockNumber := lastValidityWitness.BlockWitness.Block.BlockNumber
	depositMerkleProof, actualDepositRoot, err := d.blockBuilder.DepositTreeProof(blockNumber, depositIndex)
	if err != nil {
		return nil, common.Hash{}, err
	}
	depositTreeRoot, err := d.blockBuilder.LastDepositTreeRoot()
	if err != nil {
		return nil, common.Hash{}, err
	}

	fmt.Printf("actual deposit tree root: %s\n", actualDepositRoot.String())

	if depositTreeRoot != actualDepositRoot {
		return nil, common.Hash{}, errors.New("deposit tree root mismatch")
	}

	return depositMerkleProof, depositTreeRoot, err
}
