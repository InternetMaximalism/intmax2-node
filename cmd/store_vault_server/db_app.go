package store_vault_server

import (
	"context"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
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
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}

type BackupTransfers interface {
	CreateBackupTransfer(input *backupTransfer.UCPostBackupTransferInput) (*mDBApp.BackupTransfer, error)
	GetBackupTransfer(condition string, value string) (*mDBApp.BackupTransfer, error)
	GetBackupTransfers(condition string, value interface{}) ([]*mDBApp.BackupTransfer, error)
}

type BackupTransactions interface {
	CreateBackupTransaction(input *backupTransaction.UCPostBackupTransactionInput) (*mDBApp.BackupTransaction, error)
	GetBackupTransaction(condition string, value string) (*mDBApp.BackupTransaction, error)
	GetBackupTransactions(condition string, value interface{}) ([]*mDBApp.BackupTransaction, error)
}

type BackupDeposits interface {
	CreateBackupDeposit(input *backupDeposit.UCPostBackupDepositInput) (*mDBApp.BackupDeposit, error)
	GetBackupDeposit(condition string, value string) (*mDBApp.BackupDeposit, error)
	GetBackupDeposits(condition string, value interface{}) ([]*mDBApp.BackupDeposit, error)
}
