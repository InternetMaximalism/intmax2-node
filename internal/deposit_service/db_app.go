package deposit_service

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=deposit_service_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	EventBlockNumbers
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumber, error)
	EventBlockNumberByEventName(eventName string) (*mDBApp.EventBlockNumber, error)
	EventBlockNumbersByEventNames(eventNames []string) ([]*mDBApp.EventBlockNumber, error)
}
