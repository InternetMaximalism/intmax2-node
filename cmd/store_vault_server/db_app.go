package store_vault_server

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_db_app.go -package=store_vault_server -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommands
	BackupTransfers
	BackupTransactions
	BackupDeposits
	BackupBalances
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}

type BackupTransfers interface {
	CreateBackupTransfer(
		recipient, encryptedTransferHash, encryptedTransfer string,
		blockNumber int64,
	) (*mDBApp.BackupTransfer, error)
	GetBackupTransfer(condition string, value string) (*mDBApp.BackupTransfer, error)
	GetBackupTransfers(condition string, value interface{}) ([]*mDBApp.BackupTransfer, error)
}

type BackupTransactions interface {
	CreateBackupTransaction(
		sender, encryptedTxHash, encryptedTx, signature string,
		blockNumber int64,
	) (*mDBApp.BackupTransaction, error)
	GetBackupTransaction(condition string, value string) (*mDBApp.BackupTransaction, error)
	GetBackupTransactionBySenderAndTxDoubleHash(sender, txDoubleHash string) (*mDBApp.BackupTransaction, error)
	GetBackupTransactions(condition string, value interface{}) ([]*mDBApp.BackupTransaction, error)
}

type BackupDeposits interface {
	CreateBackupDeposit(
		recipient, encryptedDeposit string,
		blockNumber int64,
	) (*mDBApp.BackupDeposit, error)
	GetBackupDeposit(conditions []string, values []interface{}) (*mDBApp.BackupDeposit, error)
	GetBackupDeposits(condition string, value interface{}) ([]*mDBApp.BackupDeposit, error)
}

type BackupBalances interface {
	CreateBackupBalance(
		user, encryptedBalanceProof, encryptedBalanceData, signature string,
		encryptedTxs, encryptedTransfers, encryptedDeposits []string,
		blockNumber int64,
	) (*mDBApp.BackupBalance, error)
	GetBackupBalance(conditions []string, values []interface{}) (*mDBApp.BackupBalance, error)
	GetBackupBalances(condition string, value interface{}) ([]*mDBApp.BackupBalance, error)
}
