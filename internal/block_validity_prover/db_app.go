package block_validity_prover

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_validity_prover_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	Blocks
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type Blocks interface {
	UpdateBlockStatus(proposalBlockID string, status int64) error

	GetUnprocessedBlocks() ([]*mDBApp.Block, error)
}
