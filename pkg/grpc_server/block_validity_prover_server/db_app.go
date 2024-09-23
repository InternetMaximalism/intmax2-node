package block_validity_prover_server

import (
	"context"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_validity_prover_server_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommandsApp
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommandsApp interface {
	Check(ctx context.Context) health.Health
}
