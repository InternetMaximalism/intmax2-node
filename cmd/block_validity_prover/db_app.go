package block_validity_prover

import (
	"context"
	"encoding/json"
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
	Deposits
	L2BatchIndex
	RelationshipL2BatchIndexAndBlockContent
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
	BlockContentIDByL2BlockNumber(l2BlockNumber string) (bcID string, err error)
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockContentByBlockHash(blockHash string) (*mDBApp.BlockContentWithProof, error)
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
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
	DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchLastDepositIndex() (uint32, error)
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
