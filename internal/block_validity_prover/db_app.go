package block_validity_prover

import (
	"context"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_validity_prover_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}
