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
	// AccountInfo
	// DepositTreeBuilder
	// BlockHistory
	BlockContents
	EventBlockNumbersForValidityProver
	Deposits
}

type BlockBuilderStorage interface {
	GenericCommandsApp
	AccountInfo
	DepositTreeBuilder
	BlockHistory
	BlockContents
	EventBlockNumbersForValidityProver
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type AccountInfo interface {
	RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSentBlockNumber uint32) (accID uint64, err error)
	PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error)
	AccountBySenderAddress(senderAddress string) (accID *uint256.Int, err error)
}

type BlockContents interface {
	// CreateBlockContent(
	// 	blockNumber uint32,
	// 	blockHash, prevBlockHash, depositRoot, txRoot, aggregatedSignature, aggregatedPublicKey, messagePoint string,
	// 	isRegistrationBlock bool,
	// 	senders []intMaxTypes.ColumnSender,
	// ) (*mDBApp.BlockContent, error)
	CreateBlockContent(
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
	) (*mDBApp.BlockContent, error)
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContent, error)
}

type BlockHistory interface {
	GenerateBlock(blockContent *intMaxTypes.BlockContent, postedBlock *block_post_service.PostedBlock) (*BlockWitness, error)
	LatestIntMaxBlockNumber() uint32
	SetValidityProof(blockNumber uint32, proof string) error
	LastValidityProof() (*string, error)

	SetValidityWitness(blockNumber uint32, witness *ValidityWitness) error
	LastValidityWitness() (*ValidityWitness, error)
	SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
	LastSeenBlockPostedEventBlockNumber() (blockNumber uint64, err error)

	// GenerateValidityWitness(blockWitness *BlockWitness) (*ValidityWitness, error)
	NextAccountID() (uint64, error)
	AppendAccountTreeLeaf(sender *big.Int, lastBlockNumber uint64) (*intMaxTree.IndexedInsertionProof, error)
	AccountTreeRoot() (*intMaxGP.PoseidonHashOut, error)
	GetAccountTreeLeaf(sender *big.Int) (*intMaxTree.IndexedMerkleLeaf, error)
	UpdateAccountTreeLeaf(sender *big.Int, lastBlockNumber uint64) (*intMaxTree.IndexedUpdateProof, error)
	GetAccountMembershipProof(currentBlockNumber uint32, publicKey *big.Int) (*intMaxTree.IndexedMembershipProof, error)
	AppendBlockTreeLeaf(block *block_post_service.PostedBlock) error
	BlockTreeRoot() (*intMaxGP.PoseidonHashOut, error)
	BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.MerkleProof, error)
	CurrentBlockTreeProof(leafBlockNumber uint32) (*intMaxTree.MerkleProof, error)
}

type DepositTreeBuilder interface {
	LastDepositTreeRoot() (common.Hash, error)
	// AppendDepositTreeRoot(depositTreeRoot common.Hash) error
	AppendDepositTreeLeaf(depositHash common.Hash, depositLeaf *intMaxTree.DepositLeaf) (root common.Hash, err error)

	IsSynchronizedDepositIndex(depositIndex uint32) (bool, error)
	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, error)
	GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *DepositLeafWithId, depositIndex *uint32, err error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error

	FetchLastDepositIndex() (uint32, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumber, error)
	EventBlockNumberByEventName(eventName string) (*mDBApp.EventBlockNumber, error)
	EventBlockNumbersByEventNames(eventNames []string) ([]*mDBApp.EventBlockNumber, error)
}

type EventBlockNumbersForValidityProver interface {
	UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumberForValidityProver, error)
	EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error)
}

type CtrlEventBlockNumbersJobs interface {
	CreateCtrlEventBlockNumbersJobs(eventName string) error
	CtrlEventBlockNumbersJobs(eventName string) (*mDBApp.CtrlEventBlockNumbersJobs, error)
}

type EventBlockNumbersErrors interface {
	UpsertEventBlockNumbersErrors(
		eventName string,
		blockNumber *uint256.Int,
		options []byte,
		updErr error,
	) error
	EventBlockNumbersErrors(
		eventName string,
		blockNumber *uint256.Int,
	) (*mDBApp.EventBlockNumbersErrors, error)
}

type Senders interface {
	CreateSenders(
		address, publicKey string,
	) (*mDBApp.Sender, error)
	SenderByID(id string) (*mDBApp.Sender, error)
	SenderByAddress(address string) (*mDBApp.Sender, error)
	SenderByPublicKey(publicKey string) (*mDBApp.Sender, error)
}

type Accounts interface {
	CreateAccount(senderID string) (*mDBApp.Account, error)
	AccountBySenderID(senderID string) (*mDBApp.Account, error)
	AccountByAccountID(accountID *uint256.Int) (*mDBApp.Account, error)
	ResetSequenceByAccounts() error
	DelAllAccounts() error
}

type Deposits interface {
	CreateDeposit(depositLeaf intMaxTree.DepositLeaf, depositID uint32) (*mDBApp.Deposit, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	Deposit(ID string) (*mDBApp.Deposit, error)
	DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error)
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchLastDepositIndex() (uint32, error)
}
