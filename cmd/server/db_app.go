package server

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_db_app.go -package=server -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommands
	Blocks
	Deposits
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}

type Blocks interface {
	BlockByTxRoot(txRoot string) (*mDBApp.Block, error)
}

type Deposits interface {
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
}
