package server

import (
	"context"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_db_app.go -package=server -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommands
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}
