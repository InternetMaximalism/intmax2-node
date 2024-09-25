package block_validity_prover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_builder_storage"
	bbsTypes "intmax2-node/internal/block_builder_storage/types"
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
	blockBuilder block_builder_storage.BlockBuilderStorage
}

type BlockValidityProverMemory = blockValidityProver

func NewBlockValidityProver(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	db SQLDriverApp,
) (BlockValidityProver, error) {
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

	var blockBuilder block_builder_storage.BlockBuilderStorage
	blockBuilder, err = block_builder_storage.NewBlockBuilderStorage(cfg, log)
	if err != nil {
		return nil, errors.Join(ErrNewBlockBuilderStorageFail, err)
	}

	err = blockBuilder.Init(db)
	if err != nil {
		return nil, errors.Join(ErrInitBlockBuilderStorageFail, err)
	}

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

func NewBlockValidityService(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
	db SQLDriverApp,
) (BlockValidityProver, error) {
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

	var blockBuilder block_builder_storage.BlockBuilderStorage
	blockBuilder, err = block_builder_storage.NewBlockBuilderStorage(cfg, log)
	if err != nil {
		return nil, errors.Join(ErrNewBlockBuilderStorageFail, err)
	}

	err = blockBuilder.Init(db)
	if err != nil {
		return nil, errors.Join(ErrInitBlockBuilderStorageFail, err)
	}

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

func (d *blockValidityProver) FetchLastDepositIndex(db SQLDriverApp) (uint32, error) {
	return d.blockBuilder.FetchLastDepositIndex(db)
}

func (d *blockValidityProver) LastSeenBlockPostedEventBlockNumber(db SQLDriverApp) (uint64, error) {
	return d.blockBuilder.LastSeenBlockPostedEventBlockNumber(db)
}

func (d *blockValidityProver) SetLastSeenBlockPostedEventBlockNumber(db SQLDriverApp, blockNumber uint64) error {
	return d.blockBuilder.SetLastSeenBlockPostedEventBlockNumber(db, blockNumber)
}

func (d *blockValidityProver) LatestIntMaxBlockNumber() (uint32, error) {
	return d.blockBuilder.LatestIntMaxBlockNumber(), nil
}

func (d *blockValidityProver) LastPostedBlockNumber(db SQLDriverApp) (uint32, error) {
	return d.blockBuilder.LastPostedBlockNumber(db)
}

func (d *blockValidityProver) GetDepositInfoByHash(
	db SQLDriverApp,
	depositHash common.Hash,
) (*bbsTypes.DepositInfo, error) {
	depositLeafWithId, depositIndex, err := d.blockBuilder.GetDepositLeafAndIndexByHash(db, depositHash)
	if err != nil {
		var ErrGetDepositLeafAndIndexByHashFail = errors.New("failed to get deposit leaf and index by hash")
		return nil, errors.Join(ErrGetDepositLeafAndIndexByHashFail, err)
	}

	depositInfo := bbsTypes.DepositInfo{
		DepositId:    depositLeafWithId.DepositId,
		DepositIndex: depositIndex,
		DepositLeaf:  depositLeafWithId.DepositLeaf,
	}
	if depositIndex != nil {
		var blockNumber uint32
		blockNumber, err = d.blockBuilder.BlockNumberByDepositIndex(db, *depositIndex)
		if err != nil {
			var ErrBlockNumberByDepositIndexFail = errors.New("failed to get block number by deposit index")
			return nil, errors.Join(ErrBlockNumberByDepositIndexFail, err)
		}

		var isSynchronizedDepositIndex bool
		isSynchronizedDepositIndex, err = d.blockBuilder.IsSynchronizedDepositIndex(db, *depositIndex)
		if err != nil {
			var ErrIsSynchronizedDepositIndexFail = errors.New("failed to check if deposit index is synchronized")
			return nil, errors.Join(ErrIsSynchronizedDepositIndexFail, err)
		}

		depositInfo.BlockNumber = &blockNumber
		depositInfo.IsSynchronized = isSynchronizedDepositIndex
	}

	return &depositInfo, nil
}

func (d *blockValidityProver) BlockNumberByDepositIndex(db SQLDriverApp, depositIndex uint32) (uint32, error) {
	// TODO: implement this method
	return d.blockBuilder.BlockNumberByDepositIndex(db, depositIndex)
}

func (d *blockValidityProver) LatestSynchronizedBlockNumber(db SQLDriverApp) (uint32, error) {
	return d.blockBuilder.LastGeneratedProofBlockNumber(db)
}

func (d *blockValidityProver) FetchValidityProverInfo(db SQLDriverApp) (*bbsTypes.ValidityProverInfo, error) {
	lastDepositIndex, err := d.FetchLastDepositIndex(db)
	if err != nil {
		return nil, err
	}

	lastBlockNumber, err := d.LatestSynchronizedBlockNumber(db)
	if err != nil {
		return nil, err
	}

	return &bbsTypes.ValidityProverInfo{
		DepositIndex: lastDepositIndex,
		BlockNumber:  lastBlockNumber,
	}, nil
}

func (d *blockValidityProver) FetchUpdateWitness(
	db SQLDriverApp,
	publicKey *intMaxAcc.PublicKey,
	currentBlockNumber *uint32,
	targetBlockNumber uint32,
	isPrevAccountTree bool,
) (*bbsTypes.UpdateWitness, error) {
	if currentBlockNumber == nil {
		// panic("currentBlockNumber == nil")
		latestBlockNumber, err := d.blockBuilder.LastPostedBlockNumber(db)
		fmt.Printf("(FetchUpdateWitness) latestBlockNumber: %d\n", latestBlockNumber)
		if err != nil {
			var ErrLastPostedBlockNumberFail = errors.New("failed to get last posted block number")
			return nil, errors.Join(ErrLastPostedBlockNumberFail, err)
		}

		return d.blockBuilder.FetchUpdateWitness(db, publicKey, latestBlockNumber, targetBlockNumber, isPrevAccountTree)
	}

	return d.blockBuilder.FetchUpdateWitness(db, publicKey, *currentBlockNumber, targetBlockNumber, isPrevAccountTree)
}

func (d *blockValidityProver) BlockTreeProof(
	rootBlockNumber, leafBlockNumber uint32,
) (*intMaxTree.PoseidonMerkleProof, error) {
	return d.blockBuilder.BlockTreeProof(rootBlockNumber, leafBlockNumber)
}

func (d *blockValidityProver) UpdateValidityWitness(
	blockContent *intMaxTypes.BlockContent,
	prevValidityWitness *bbsTypes.ValidityWitness,
) (*bbsTypes.ValidityWitness, error) {
	return d.blockBuilder.UpdateValidityWitness(blockContent, prevValidityWitness)
}

func (d *blockValidityProver) ValidityWitness(
	db SQLDriverApp,
	txRoot common.Hash,
) (*bbsTypes.ValidityWitness, error) {
	rawBlockContent, err := d.blockBuilder.BlockContentByTxRoot(db, txRoot)
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

	lastValidityWitness, err := d.blockBuilder.LastValidityWitness(db)
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

	return d.blockBuilder.CalculateValidityWitness(blockWitness)
}

// TODO: multiple response
func (d *blockValidityProver) BlockContentByTxRoot(db SQLDriverApp, txRoot common.Hash) (*block_post_service.PostedBlock, error) {
	blockContent, err := d.blockBuilder.BlockContentByTxRoot(db, txRoot)
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

func (d *blockValidityProver) ValidityPublicInputs(
	db SQLDriverApp,
	txRoot common.Hash,
) (*bbsTypes.ValidityPublicInputs, []bbsTypes.SenderLeaf, error) {
	validityWitness, err := d.ValidityWitness(db, txRoot)
	if err != nil {
		return nil, nil, err
	}

	validityPublicInputs := validityWitness.ValidityPublicInputs(d.log)
	senderLeaves := validityWitness.ValidityTransitionWitness.SenderLeaves

	return validityPublicInputs, senderLeaves, nil
}

func (d *blockValidityProver) DepositTreeProof(
	db SQLDriverApp,
	depositIndex uint32,
) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	validityProverInfo, err := d.FetchValidityProverInfo(db)
	if err != nil {
		return nil, common.Hash{}, err
	}

	latestBlockNumber := validityProverInfo.BlockNumber
	var (
		depositMerkleProof *intMaxTree.KeccakMerkleProof
		actualDepositRoot  common.Hash
	)
	depositMerkleProof, actualDepositRoot, err = d.blockBuilder.DepositTreeProof(latestBlockNumber, depositIndex)
	if err != nil {
		return nil, common.Hash{}, err
	}

	var depositTreeRoot common.Hash
	depositTreeRoot, err = d.blockBuilder.LastDepositTreeRoot()
	if err != nil {
		return nil, common.Hash{}, err
	}

	fmt.Printf("actual deposit tree root: %s\n", actualDepositRoot.String())

	if depositTreeRoot != actualDepositRoot {
		return nil, common.Hash{}, errors.New("deposit tree root mismatch")
	}

	return depositMerkleProof, depositTreeRoot, err
}
