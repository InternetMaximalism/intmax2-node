package block_validity_prover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/utils"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func NewBlockValidityProver(ctx context.Context, cfg *configs.Config, log logger.Logger, sb ServiceBlockchain, db SQLDriverApp) (BlockValidityProver, error) {
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

// func (d *blockValidityProver) RollupContractDeployedBlockNumber() (uint64, error) {
// 	return d.cfg.Blockchain.RollupContractDeployedBlockNumber, nil
// }

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

func NewBlockValidityService(ctx context.Context, cfg *configs.Config, log logger.Logger, sb ServiceBlockchain, db SQLDriverApp) (*blockValidityProver, error) {
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

func (d *blockValidityProver) LatestIntMaxBlockNumber() (uint32, error) {
	return d.blockBuilder.LatestIntMaxBlockNumber(), nil
}

type DepositInfo struct {
	DepositId      uint32
	DepositIndex   *uint32
	BlockNumber    *uint32
	IsSynchronized bool
	DepositLeaf    *intMaxTree.DepositLeaf
}

func (d *blockValidityProver) GetDepositInfoByHash(depositHash common.Hash) (*DepositInfo, error) {
	depositLeafWithId, depositIndex, err := d.blockBuilder.GetDepositLeafAndIndexByHash(depositHash)
	if err != nil {
		var ErrGetDepositLeafAndIndexByHashFail = errors.New("failed to get deposit leaf and index by hash")
		return nil, errors.Join(ErrGetDepositLeafAndIndexByHashFail, err)
	}

	depositInfo := DepositInfo{
		DepositId:    depositLeafWithId.DepositId,
		DepositIndex: depositIndex,
		DepositLeaf:  depositLeafWithId.DepositLeaf,
	}
	if depositIndex != nil {
		blockNumber, err := d.blockBuilder.BlockNumberByDepositIndex(*depositIndex)
		if err != nil {
			var ErrBlockNumberByDepositIndexFail = errors.New("failed to get block number by deposit index")
			return nil, errors.Join(ErrBlockNumberByDepositIndexFail, err)
		}

		isSynchronizedDepositIndex, err := d.blockBuilder.IsSynchronizedDepositIndex(*depositIndex)
		if err != nil {
			var ErrIsSynchronizedDepositIndexFail = errors.New("failed to check if deposit index is synchronized")
			return nil, errors.Join(ErrIsSynchronizedDepositIndexFail, err)
		}

		depositInfo.BlockNumber = &blockNumber
		depositInfo.IsSynchronized = isSynchronizedDepositIndex
	}

	return &depositInfo, nil
}

func (d *blockValidityProver) BlockNumberByDepositIndex(depositIndex uint32) (uint32, error) {
	// TODO: implement this method
	return d.blockBuilder.BlockNumberByDepositIndex(depositIndex)
}

func (d *blockValidityProver) LatestSynchronizedBlockNumber() (uint32, error) {
	return d.blockBuilder.LastGeneratedProofBlockNumber()
}

type ValidityProverInfo struct {
	DepositIndex uint32
	BlockNumber  uint32
}

func (d *blockValidityProver) FetchValidityProverInfo() (*ValidityProverInfo, error) {
	lastDepositIndex, err := d.FetchLastDepositIndex()
	if err != nil {
		return nil, err
	}

	lastBlockNumber, err := d.LatestSynchronizedBlockNumber()
	if err != nil {
		return nil, err
	}

	return &ValidityProverInfo{
		DepositIndex: lastDepositIndex,
		BlockNumber:  lastBlockNumber,
	}, nil
}

func (d *blockValidityProver) FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber *uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error) {
	if currentBlockNumber == nil {
		// panic("currentBlockNumber == nil")
		latestBlockNumber, err := d.blockBuilder.db.LastPostedBlockNumber()
		fmt.Printf("(FetchUpdateWitness) latestBlockNumber: %d\n", latestBlockNumber)
		if err != nil {
			var ErrLastPostedBlockNumberFail = errors.New("failed to get last posted block number")
			return nil, errors.Join(ErrLastPostedBlockNumberFail, err)
		}

		return d.blockBuilder.FetchUpdateWitness(publicKey, latestBlockNumber, targetBlockNumber, isPrevAccountTree)
	}

	return d.blockBuilder.FetchUpdateWitness(publicKey, *currentBlockNumber, targetBlockNumber, isPrevAccountTree)
}

func (d *blockValidityProver) BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.PoseidonMerkleProof, error) {
	return d.blockBuilder.BlockTreeProof(rootBlockNumber, leafBlockNumber)
}

func (d *blockValidityProver) UpdateValidityWitness(
	blockContent *intMaxTypes.BlockContent,
	prevValidityWitness *ValidityWitness,
) (*ValidityWitness, error) {
	return d.blockBuilder.UpdateValidityWitness(blockContent, prevValidityWitness)
}

func (d *blockValidityProver) ValidityWitness(
	txRoot common.Hash,
) (*ValidityWitness, error) {
	rawBlockContent, err := d.blockBuilder.BlockContentByTxRoot(txRoot)
	if err != nil {
		return nil, err
	}

	senderType := intMaxTypes.AccountIDSenderType
	if rawBlockContent.IsRegistrationBlock {
		senderType = intMaxTypes.PublicKeySenderType
	}

	var senders []intMaxTypes.Sender
	err = json.Unmarshal(rawBlockContent.Senders, &senders)
	if err != nil {
		var ErrUnmarshalSendersFail = errors.New("failed to unmarshal senders")
		return nil, errors.Join(ErrUnmarshalSendersFail, err)
	}

	aggregatedSignatureBytes, err := hexutil.Decode(rawBlockContent.AggregatedSignature)
	if err != nil {
		var ErrDecodeAggregatedSignatureFail = errors.New("failed to decode aggregated signature")
		return nil, errors.Join(ErrDecodeAggregatedSignatureFail, err)
	}

	aggregatedSignature := new(bn254.G2Affine)
	err = aggregatedSignature.Unmarshal(aggregatedSignatureBytes)
	if err != nil {
		var ErrUnmarshalAggregatedSignatureFail = errors.New("failed to unmarshal aggregated signature")
		return nil, errors.Join(ErrUnmarshalAggregatedSignatureFail, err)
	}

	blockContent := intMaxTypes.NewBlockContent(
		senderType,
		senders,
		common.HexToHash(rawBlockContent.TxRoot),
		aggregatedSignature,
	)

	lastValidityWitness, err := d.blockBuilder.LastValidityWitness()
	if err != nil {
		var ErrLastValidityWitnessNotFound = errors.New("last validity witness not found")
		return nil, errors.Join(ErrLastValidityWitnessNotFound, err)
	}
	blockWitness, err := d.blockBuilder.GenerateBlockWithTxTreeFromBlockContent(
		blockContent,
		lastValidityWitness.BlockWitness.Block,
	)
	if err != nil {
		panic(err)
	}

	return calculateValidityWitness(d.blockBuilder, blockWitness)
}

// TODO: multiple response
func (d *blockValidityProver) BlockContentByTxRoot(txRoot common.Hash) (*block_post_service.PostedBlock, error) {
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

func (d *blockValidityProver) ValidityPublicInputs(txRoot common.Hash) (*ValidityPublicInputs, []SenderLeaf, error) {
	validityWitness, err := d.ValidityWitness(txRoot)
	if err != nil {
		return nil, nil, err
	}

	validityPublicInputs := validityWitness.ValidityPublicInputs()
	senderLeaves := validityWitness.ValidityTransitionWitness.SenderLeaves

	return validityPublicInputs, senderLeaves, nil
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
