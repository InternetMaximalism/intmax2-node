package block_validity_prover

import (
	"intmax2-node/internal/block_builder_storage"
	bbsTypes "intmax2-node/internal/block_builder_storage/types"
	intMaxTypes "intmax2-node/internal/types"

	// intMaxAcc "intmax2-node/internal/accounts"

	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

type BlockValidityService interface {
	BlockContentByTxRoot(db SQLDriverApp, txRoot common.Hash) (*block_post_service.PostedBlock, error)
	GetDepositInfoByHash(db SQLDriverApp, depositHash common.Hash) (depositInfo *bbsTypes.DepositInfo, err error)
	FetchValidityProverInfo(db SQLDriverApp) (*bbsTypes.ValidityProverInfo, error)
	FetchUpdateWitness(db SQLDriverApp, publicKey *intMaxAcc.PublicKey, currentBlockNumber *uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*bbsTypes.UpdateWitness, error)
	DepositTreeProof(db SQLDriverApp, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.PoseidonMerkleProof, error)
	ValidityPublicInputs(
		db SQLDriverApp,
		txRoot common.Hash,
	) (
		validityPublicInputs *bbsTypes.ValidityPublicInputs,
		senderLeaves []bbsTypes.SenderLeaf,
		err error,
	)
}

type BlockValidityProver interface {
	FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error)
	FetchLastDepositIndex(db SQLDriverApp) (uint32, error)
	LastSeenBlockPostedEventBlockNumber(db SQLDriverApp) (uint64, error)
	SetLastSeenBlockPostedEventBlockNumber(db SQLDriverApp, blockNumber uint64) error
	LatestIntMaxBlockNumber() (uint32, error)
	LastPostedBlockNumber(db SQLDriverApp) (uint32, error)
	GetDepositInfoByHash(
		db SQLDriverApp,
		depositHash common.Hash,
	) (*bbsTypes.DepositInfo, error)
	BlockNumberByDepositIndex(db SQLDriverApp, depositIndex uint32) (uint32, error)
	LatestSynchronizedBlockNumber(db SQLDriverApp) (uint32, error)
	FetchValidityProverInfo(db SQLDriverApp) (*bbsTypes.ValidityProverInfo, error)
	FetchUpdateWitness(
		db SQLDriverApp,
		publicKey *intMaxAcc.PublicKey,
		currentBlockNumber *uint32,
		targetBlockNumber uint32,
		isPrevAccountTree bool,
	) (*bbsTypes.UpdateWitness, error)
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (*intMaxTree.PoseidonMerkleProof, error)
	UpdateValidityWitness(
		blockContent *intMaxTypes.BlockContent,
		prevValidityWitness *bbsTypes.ValidityWitness,
	) (*bbsTypes.ValidityWitness, error)
	ValidityWitness(
		db SQLDriverApp,
		txRoot common.Hash,
	) (*bbsTypes.ValidityWitness, error)
	BlockContentByTxRoot(db SQLDriverApp, txRoot common.Hash) (*block_post_service.PostedBlock, error)
	ValidityPublicInputs(
		db SQLDriverApp,
		txRoot common.Hash,
	) (*bbsTypes.ValidityPublicInputs, []bbsTypes.SenderLeaf, error)
	DepositTreeProof(
		db SQLDriverApp,
		depositIndex uint32,
	) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockBuilder() block_builder_storage.BlockBuilderStorage
	SyncBlockTree(db SQLDriverApp, bps BlockSynchronizer, startBlock uint64) (lastEventSeenBlockNumber uint64, err error)
	SyncBlockProverWithBlockNumber(
		db SQLDriverApp,
		blockNumber uint32,
	) error
	SyncDepositTree(db SQLDriverApp, latestBlock *uint64, depositIndex uint32) error
	SyncDepositedEvents(db SQLDriverApp) error
}
