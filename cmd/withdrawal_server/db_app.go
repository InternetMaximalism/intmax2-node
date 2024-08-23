package withdrawal_server

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_db_app.go -package=withdrawal_server -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommands
	Withdrawals
	EventBlockNumbers
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}

type Withdrawals interface {
	CreateWithdrawal(
		id string,
		transferData *mDBApp.TransferData,
		transferMerkleProof *mDBApp.TransferMerkleProof,
		transaction *mDBApp.Transaction,
		txMerkleProof *mDBApp.TxMerkleProof,
		transferHash string,
		blockNumber int64,
		blockHash string,
		enoughBalanceProof *mDBApp.EnoughBalanceProof,
	) (*mDBApp.Withdrawal, error)
	UpdateWithdrawalsStatus(ids []string, status mDBApp.WithdrawalStatus) error
	WithdrawalByID(id string) (*mDBApp.Withdrawal, error)
	WithdrawalsByHashes(transferHashes []string) (*[]mDBApp.Withdrawal, error)
	WithdrawalsByStatus(status mDBApp.WithdrawalStatus, limit *int) (*[]mDBApp.Withdrawal, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumber, error)
	EventBlockNumberByEventName(eventName string) (*mDBApp.EventBlockNumber, error)
}
