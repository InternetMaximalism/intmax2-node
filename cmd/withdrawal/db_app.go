package withdrawal

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=mock_db_app.go -package=withdrawal -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	Withdrawals
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type Withdrawals interface {
	CreateWithdrawal(w *mDBApp.Withdrawal) (*mDBApp.Withdrawal, error)
	FindWithdrawals(status mDBApp.WithdrawalStatus) (*[]mDBApp.Withdrawal, error)
}
