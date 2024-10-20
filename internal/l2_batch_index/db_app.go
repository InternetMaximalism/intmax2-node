package l2_batch_index

import (
	"context"
	"encoding/json"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=l2_batch_index_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	CtrlProcessingJobs
	L2BatchIndex
	BlockContents
	RelationshipL2BatchIndexAndBlockContent
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type CtrlProcessingJobs interface {
	CreateCtrlProcessingJobs(name string, options json.RawMessage) error
	CtrlProcessingJobsByMaskName(mask string) (*mDBApp.CtrlProcessingJobs, error)
	UpdatedAtOfCtrlProcessingJobByName(name string, updatedAt time.Time) (err error)
	DeleteCtrlProcessingJobByName(name string) (err error)
}

type L2BatchIndex interface {
	CreateL2BatchIndex(batchIndex *uint256.Int) (err error)
	L2BatchIndex(batchIndex *uint256.Int) (*mDBApp.L2BatchIndex, error)
	UpdOptionsOfBatchIndex(batchIndex *uint256.Int, options json.RawMessage) (err error)
	UpdL1VerifiedBatchTxHashOfBatchIndex(batchIndex *uint256.Int, hash string) (err error)
}

type BlockContents interface {
	BlockContentIDByL2BlockNumber(l2BlockNumber string) (bcID string, err error)
}

type RelationshipL2BatchIndexAndBlockContent interface {
	CreateRelationshipL2BatchIndexAndBlockContentID(
		batchIndex *uint256.Int,
		blockContentID string,
	) (err error)
}
