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
	intMaxTypes "intmax2-node/internal/types"
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

func (d *blockValidityProver) FetchNextDepositIndex() (uint32, error) {
	nextDepositIndex, err := d.blockBuilder.FetchNextDepositIndex()
	if err != nil {
		return 0, err
	}

	return nextDepositIndex, nil
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

func (d *blockValidityProver) LastWitnessGeneratedBlockNumber() (uint32, error) {
	return d.blockBuilder.LastWitnessGeneratedBlockNumber(), nil
}

func (d *blockValidityProver) LastPostedBlockNumber() (uint32, error) {
	return d.blockBuilder.db.LastPostedBlockNumber()
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
		startIntMaxBlockNumber := uint32(1)

		var blockNumber uint32
		blockNumber, err = d.blockBuilder.BlockNumberByDepositIndex(*depositIndex, &startIntMaxBlockNumber)
		if err != nil {
			var ErrBlockNumberByDepositIndexFail = errors.New("failed to get block number by deposit index")
			return nil, errors.Join(ErrBlockNumberByDepositIndexFail, err)
		}

		var isSynchronizedDepositIndex bool
		isSynchronizedDepositIndex, err = d.blockBuilder.IsSynchronizedDepositIndex(*depositIndex)
		if err != nil {
			var ErrIsSynchronizedDepositIndexFail = errors.New("failed to check if deposit index is synchronized")
			return nil, errors.Join(ErrIsSynchronizedDepositIndexFail, err)
		}

		depositInfo.BlockNumber = &blockNumber
		depositInfo.IsSynchronized = isSynchronizedDepositIndex
	}

	return &depositInfo, nil
}

// func (d *blockValidityProver) BlockNumberByDepositIndex(depositIndex uint32) (uint32, error) {
// 	// TODO: implement this method
// 	return d.blockBuilder.BlockNumberByDepositIndex(depositIndex)
// }

func (d *blockValidityProver) LatestSynchronizedBlockNumber() (uint32, error) {
	return d.blockBuilder.LastGeneratedProofBlockNumber()
}

type ValidityProverInfo struct {
	DepositIndex uint32
	BlockNumber  uint32
}

func (d *blockValidityProver) FetchValidityProverInfo() (*ValidityProverInfo, error) {
	nextDepositIndex, err := d.FetchNextDepositIndex()
	if err != nil {
		return nil, err
	}

	lastBlockNumber, err := d.LatestSynchronizedBlockNumber()
	if err != nil {
		return nil, err
	}

	return &ValidityProverInfo{
		DepositIndex: nextDepositIndex,
		BlockNumber:  lastBlockNumber,
	}, nil
}

// Returns an update witness for a given public key and block numbers.
// If the current block number is not provided, it fetches the latest posted block number from the database.
// It then uses this block number, or the provided current block number, to fetch the update witness
// from the block builder. The function returns the update witness or an error if the operation fails.
func (d *blockValidityProver) FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error) {
	return d.blockBuilder.FetchUpdateWitness(publicKey, currentBlockNumber, targetBlockNumber, isPrevAccountTree)
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
		return nil, fmt.Errorf("failed to get block content by tx root: %w", err)
	}

	auxInfo, err := blockAuxInfoFromBlockContent(rawBlockContent)
	if err != nil {
		return nil, fmt.Errorf("failed to get block aux info from block content: %w", err)
	}
	blockWitness, err := d.blockBuilder.GenerateBlockWithTxTreeFromBlockContentAndPrevBlock(
		auxInfo.BlockContent,
		auxInfo.PostedBlock,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate block witness: %w", err)
	}

	fmt.Printf("(validityWitness) blockWitness.AccountMembershipProofs: %v\n", blockWitness.AccountMembershipProofs.IsSome)
	validityWitness, _, _, err := calculateValidityWitness(d.blockBuilder, blockWitness)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate validity witness: %w", err)
	}

	return validityWitness, err
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

func (d *blockValidityProver) LatestDepositTreeProofByBlockNumber(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	validityProverInfo, err := d.FetchValidityProverInfo()
	if err != nil {
		return nil, common.Hash{}, err
	}

	latestBlockNumber := validityProverInfo.BlockNumber

	return d.DepositTreeProof(latestBlockNumber, depositIndex)
}

func (d *blockValidityProver) DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error) {
	depositMerkleProof, actualDepositRoot, err := d.blockBuilder.DepositTreeProof(blockNumber, depositIndex)
	if err != nil {
		return nil, common.Hash{}, err
	}
	d.log.Debugf("actual deposit tree root: %s\n", actualDepositRoot.String())
	// depositTreeRoot, err := d.blockBuilder.LastDepositTreeRoot()
	// if err != nil {
	// 	return nil, common.Hash{}, err
	// }
	// if depositTreeRoot != actualDepositRoot {
	// 	d.log.Debugf("expected deposit tree root: %s\n", depositTreeRoot.String())
	// 	return nil, common.Hash{}, errors.New("deposit tree root mismatch")
	// }

	// debug
	depositLeaf := d.blockBuilder.MerkleTreeHistory.MerkleTrees[blockNumber].DepositLeaves[depositIndex]
	fmt.Printf("depositIndex: %+v\n", depositIndex)
	fmt.Printf("depositLeaf: %+v\n", depositLeaf)
	fmt.Printf("depositLeaf RecipientSaltHash: %v\n", common.Hash(depositLeaf.RecipientSaltHash).String())
	fmt.Printf("depositLeaf hash: %s\n", depositLeaf.Hash().String())
	err = depositMerkleProof.Verify(depositLeaf.Hash(), int(depositIndex), actualDepositRoot)
	if err != nil {
		panic(err)
	}

	return depositMerkleProof, actualDepositRoot, err
}
