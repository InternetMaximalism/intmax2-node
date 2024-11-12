package block_validity_prover

import (
	"context"
	"encoding/json"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app.go -package=block_validity_prover -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommands
	CtrlProcessingJobs
	BlockContents
	EventBlockNumbersForValidityProver
	Senders
	Accounts
	BlockContainedSenders
	Deposits
	L2BatchIndex
	RelationshipL2BatchIndexAndBlockContent
	EthereumCounterparties
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}

type CtrlProcessingJobs interface {
	CreateCtrlProcessingJobs(name string, options json.RawMessage) error
	CtrlProcessingJobsByMaskName(mask string) (*mDBApp.CtrlProcessingJobs, error)
	UpdatedAtOfCtrlProcessingJobByName(name string, updatedAt time.Time) (err error)
	DeleteCtrlProcessingJobByName(name string) (err error)
}

type BlockContents interface {
	CreateBlockContent(
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
		l2BlockNumber *uint256.Int,
		l2BlockHash common.Hash,
	) (*mDBApp.BlockContentWithProof, error)
	BlockContentUpdDepositLeavesCounterByBlockNumber(
		blockNumber, depositLeavesCounter uint32,
	) error
	BlockContentIDByL2BlockNumber(l2BlockNumber string) (bcID string, err error)
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockContentByBlockHash(blockHash string) (*mDBApp.BlockContentWithProof, error)
	BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error)
	BlockContentListByTxRoot(txRoot ...common.Hash) ([]*mDBApp.BlockContentWithProof, error)
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

type Senders interface {
	CreateSenders(
		address, publicKey string,
	) (*mDBApp.Sender, error)
	SenderByID(id string) (*mDBApp.Sender, error)
	SenderByAddress(address string) (*mDBApp.Sender, error)
	// SenderByPublicKey(publicKey string) (*mDBApp.Sender, error)
}

type Accounts interface {
	CreateAccount(senderID string) (*mDBApp.Account, error)
	AccountBySender(publicKey *intMaxAcc.PublicKey) (*mDBApp.Account, error)
	AccountBySenderID(senderID string) (*mDBApp.Account, error)
}

type BlockContainedSenders interface {
	CreateBlockParticipant(
		blockNumber uint32,
		senderId string,
	) (*mDBApp.BlockContainedSender, error)
}

type Deposits interface {
	CreateDeposit(depositLeaf intMaxTree.DepositLeaf, depositID uint32, sender string) (*mDBApp.Deposit, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	UpdateSenderByDepositID(depositID uint32, sender string) error
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
	DepositsListByDepositHash(depositHash ...common.Hash) ([]*mDBApp.Deposit, error)
	DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchNextDepositIndex() (uint32, error)
}

type L2BatchIndex interface {
	CreateL2BatchIndex(batchIndex *uint256.Int) (err error)
	L2BatchIndex(batchIndex *uint256.Int) (*mDBApp.L2BatchIndex, error)
	UpdOptionsOfBatchIndex(batchIndex *uint256.Int, options json.RawMessage) (err error)
	UpdL1VerifiedBatchTxHashOfBatchIndex(batchIndex *uint256.Int, hash string) (err error)
}

type RelationshipL2BatchIndexAndBlockContent interface {
	CreateRelationshipL2BatchIndexAndBlockContentID(
		batchIndex *uint256.Int,
		blockContentID string,
	) (err error)
	RelationshipL2BatchIndexAndBlockContentsByBlockContentID(
		blockContentID string,
	) (*mDBApp.RelationshipL2BatchIndexBlockContents, error)
}

type EthereumCounterparties interface {
	CreateEthereumCounterparty(
		address string,
	) (*mDBApp.EthereumCounterparty, error)
	EthereumCounterpartyByAddress(address string) (*mDBApp.EthereumCounterparty, error)
}
