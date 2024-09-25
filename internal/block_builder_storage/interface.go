package block_builder_storage

import (
	intMaxAcc "intmax2-node/internal/accounts"
	bbsTypes "intmax2-node/internal/block_builder_storage/types"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type BlockBuilderStorage interface {
	Init(db SQLDriverApp) (err error)
	FetchUpdateWitness(
		db SQLDriverApp,
		publicKey *intMaxAcc.PublicKey,
		currentBlockNumber uint32,
		targetBlockNumber uint32,
		isPrevAccountTree bool,
	) (*bbsTypes.UpdateWitness, error)
	ValidityProofByBlockNumber(db SQLDriverApp, blockNumber uint32) (*string, error)
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (*intMaxTree.PoseidonMerkleProof, error)
	GetAccountMembershipProof(
		blockNumber uint32,
		publicKey *big.Int,
	) (*intMaxTree.IndexedMembershipProof, error)
	GenerateBlock(
		blockContent *intMaxTypes.BlockContent,
		postedBlock *block_post_service.PostedBlock,
	) (*bbsTypes.BlockWitness, error)
	FetchLastDepositIndex(db SQLDriverApp) (uint32, error)
	LatestIntMaxBlockNumber() uint32
	LastPostedBlockNumber(db SQLDriverApp) (uint32, error)
	EventBlockNumberByEventNameForValidityProver(
		db SQLDriverApp,
	) (*mDBApp.EventBlockNumberForValidityProver, error)
	LastSeenBlockPostedEventBlockNumber(db SQLDriverApp) (uint64, error)
	SetLastSeenBlockPostedEventBlockNumber(db SQLDriverApp, blockNumber uint64) error
	GetDepositLeafAndIndexByHash(
		db SQLDriverApp,
		depositHash common.Hash,
	) (depositLeafWithId *bbsTypes.DepositLeafWithId, depositIndex *uint32, err error)
	BlockNumberByDepositIndex(
		db SQLDriverApp,
		depositIndex uint32,
	) (uint32, error)
	SetValidityWitness(blockNumber uint32, witness *bbsTypes.ValidityWitness) error
	LastValidityWitness(db SQLDriverApp) (*bbsTypes.ValidityWitness, error)
	ValidityWitnessByBlockNumber(db SQLDriverApp, blockNumber uint32) (*bbsTypes.ValidityWitness, error)
	BlockAuxInfo(db SQLDriverApp, blockNumber uint32) (*bbsTypes.AuxInfo, error)
	GenerateBlockWithTxTreeFromBlockContent(
		blockContent *intMaxTypes.BlockContent,
		postedBlock *block_post_service.PostedBlock,
	) (*bbsTypes.BlockWitness, error)
	AppendAccountTreeLeaf(
		sender *big.Int,
		lastBlockNumber uint32,
	) (*intMaxTree.IndexedInsertionProof, error)
	UpdateAccountTreeLeaf(
		sender *big.Int,
		lastBlockNumber uint32,
	) (*intMaxTree.IndexedUpdateProof, error)
	GetAccountTreeLeaf(sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error)
	ProveInclusion(accountId uint64) (*bbsTypes.AccountMerkleProof, error)
	BlockTreeRoot(blockNumber *uint32) (*intMaxGP.PoseidonHashOut, error)
	LastGeneratedProofBlockNumber(db SQLDriverApp) (uint32, error)
	IsSynchronizedDepositIndex(db SQLDriverApp, depositIndex uint32) (bool, error)
	UpdateValidityWitness(
		blockContent *intMaxTypes.BlockContent,
		prevValidityWitness *bbsTypes.ValidityWitness,
	) (*bbsTypes.ValidityWitness, error)
	AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error)
	AppendBlockTreeLeaf(
		block *block_post_service.PostedBlock,
	) (blockNumber uint32, err error)
	CreateBlockContent(
		db SQLDriverApp,
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
	) (*mDBApp.BlockContentWithProof, error)
	BlockContentByTxRoot(db SQLDriverApp, txRoot common.Hash) (*mDBApp.BlockContentWithProof, error)
	UpdateDepositIndexByDepositHash(
		db SQLDriverApp,
		depositHash common.Hash,
		depositIndex uint32,
	) error
	UpsertEventBlockNumberForValidityProver(
		db SQLDriverApp,
		eventName string,
		blockNumber uint64,
	) (*mDBApp.EventBlockNumberForValidityProver, error)
	AppendDepositTreeLeaf(
		depositHash common.Hash,
		depositLeaf *intMaxTree.DepositLeaf,
	) (root common.Hash, nextIndex uint32, err error)
	DepositTreeProof(
		blockNumber, depositIndex uint32,
	) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	LastDepositTreeRoot() (common.Hash, error)
	NextAccountID() (uint64, error)
	SetValidityProof(
		db SQLDriverApp,
		blockHash common.Hash,
		proof string,
	) error
	RegisterPublicKey(
		pk *intMaxAcc.PublicKey,
		lastSeenBlockNumber uint32,
	) (accountID uint64, err error)
	PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error)
	AccountBySenderAddress(_ string) (*uint256.Int, error)
	CalculateValidityWitness(blockWitness *bbsTypes.BlockWitness) (*bbsTypes.ValidityWitness, error)
	CalculateValidityWitnessWithConsistencyCheck(
		blockWitness *bbsTypes.BlockWitness,
		prevValidityWitness *bbsTypes.ValidityWitness,
	) (*bbsTypes.ValidityWitness, error)
}
