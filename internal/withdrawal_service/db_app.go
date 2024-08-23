package withdrawal_service

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=withdrawal_service_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	Withdrawals
	EventBlockNumbers
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
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
