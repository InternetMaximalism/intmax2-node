package block_validity_prover

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_validity_prover_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	BlockContents
	EventBlockNumbersForValidityProver
	Deposits
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type BlockContents interface {
	CreateBlockContent(
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
	) (*mDBApp.BlockContentWithProof, error)
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error)
	ScanBlockHashAndSenders() (blockHashAndSendersMap map[uint32]mDBApp.BlockHashAndSenders, lastBlockNumber uint32, err error)
	CreateValidityProof(blockHash common.Hash, validityProof []byte) (*mDBApp.BlockProof, error)
	LastBlockValidityProof() (*mDBApp.BlockContentWithProof, error)
	LastBlockNumberGeneratedValidityProof() (uint32, error)
	LastPostedBlockNumber() (uint32, error)
}

type EventBlockNumbersForValidityProver interface {
	UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumberForValidityProver, error)
	EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error)
}

type Deposits interface {
	CreateDeposit(depositLeaf intMaxTree.DepositLeaf, depositID uint32) (*mDBApp.Deposit, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	// Deposit(ID string) (*mDBApp.Deposit, error)
	// DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error)
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchLastDepositIndex() (uint32, error)
}

type BlockBuilderStorage interface {
	GenericCommandsApp
	AccountInfo
	BuilderBlockContents
	BlockHistory
	BuilderDeposits
	// DepositTreeBuilder
	// BuilderEventBlockNumbersForValidityProver
}

type AccountInfo interface {
	// RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSentBlockNumber uint32) (accID uint64, err error)
	PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error)
	AccountBySenderAddress(senderAddress string) (accID *uint256.Int, err error)
}

type BuilderBlockContents interface {
	// CreateBlockContent(
	// 	postedBlock *block_post_service.PostedBlock,
	// 	blockContent *intMaxTypes.BlockContent,
	// ) (*mDBApp.BlockContentWithProof, error)
	// BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	// BlockContentByTxRoot(txRoot string) (*mDBApp.BlockContent, error)
}

type BlockHistory interface {
	// GenerateBlock(blockContent *intMaxTypes.BlockContent, postedBlock *block_post_service.PostedBlock) (*BlockWitness, error)
	LatestIntMaxBlockNumber() uint32
	// SetValidityProof(blockNumber uint32, proof string) error
	// ValidityProofByBlockNumber(blockNumber uint32) (*string, error)

	SetValidityWitness(blockNumber uint32, witness *ValidityWitness) error
	LastValidityWitness() (*ValidityWitness, error)
	// SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
	// LastSeenBlockPostedEventBlockNumber() (blockNumber uint64, err error)

	// GenerateValidityWitness(blockWitness *BlockWitness) (*ValidityWitness, error)
	// NextAccountID() (uint64, error)
	AppendAccountTreeLeaf(sender *big.Int, lastBlockNumber uint32) (*intMaxTree.IndexedInsertionProof, error)
	AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error)
	GetAccountTreeLeaf(sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error)
	UpdateAccountTreeLeaf(sender *big.Int, lastBlockNumber uint32) (*intMaxTree.IndexedUpdateProof, error)
	// GetAccountMembershipProof(currentBlockNumber uint32, publicKey *big.Int) (*intMaxTree.IndexedMembershipProof, error)
	AppendBlockTreeLeaf(block *block_post_service.PostedBlock) (uint32, error)
	BlockTreeRoot(blockNumber *uint32) (*intMaxGP.PoseidonHashOut, error)
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (
		*intMaxTree.PoseidonMerkleProof,
		*intMaxTree.PoseidonHashOut,
		error,
	)
	// CurrentBlockTreeProof(leafBlockNumber uint32) (*intMaxTree.MerkleProof, error)
}

type BuilderDeposits interface {
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
}

// type DepositTreeBuilder interface {
// 	LastDepositTreeRoot() (common.Hash, error)
// 	AppendDepositTreeLeaf(depositHash common.Hash, depositLeaf *intMaxTree.DepositLeaf) (root common.Hash, err error)

// 	IsSynchronizedDepositIndex(depositIndex uint32) (bool, error)
// 	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
// 	GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *DepositLeafWithId, depositIndex *uint32, err error)

// 	FetchLastDepositIndex() (uint32, error)
// }

// type BuilderEventBlockNumbersForValidityProver interface {
// 	UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumberForValidityProver, error)
// 	EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error)
// }
