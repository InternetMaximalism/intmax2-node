package block_validity_prover

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

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
	PublicKeyByAccountID(blockNumber uint32, accountID uint64) (pk *intMaxAcc.PublicKey, err error)
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
	LastWitnessGeneratedBlockNumber() uint32
	// SetValidityProof(blockNumber uint32, proof string) error
	// ValidityProofByBlockNumber(blockNumber uint32) (*string, error)

	SetValidityWitness(blockNumber uint32, witness *ValidityWitness, newAccountTree *intMaxTree.AccountTree, newBlockTree *intMaxTree.BlockHashTree) error
	LastValidityWitness() (*ValidityWitness, error)
	// SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
	// LastSeenBlockPostedEventBlockNumber() (blockNumber uint64, err error)

	// GenerateValidityWitness(blockWitness *BlockWitness) (*ValidityWitness, error)
	NextAccountID(blockNumber uint32) (uint64, error)
	// GetAccountMembershipProof(currentBlockNumber uint32, publicKey *big.Int) (*intMaxTree.IndexedMembershipProof, error)
	AppendBlockTreeLeaf(block *block_post_service.PostedBlock) (uint32, error)
	BlockTreeRoot(blockNumber *uint32) (*intMaxGP.PoseidonHashOut, error)
	BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.PoseidonMerkleProof, error)
	// CurrentBlockTreeProof(leafBlockNumber uint32) (*intMaxTree.MerkleProof, error)
	CopyAccountTree(dst *intMaxTree.AccountTree, blockNumber uint32) error
	CopyBlockHashTree(dst *intMaxTree.BlockHashTree, blockNumber uint32) error
}

type BuilderDeposits interface {
	ScanDeposits() ([]*mDBApp.Deposit, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
}
