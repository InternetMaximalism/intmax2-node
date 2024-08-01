package deposit_synchronizer

import (
	"context"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=deposit_synchronizer_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}
