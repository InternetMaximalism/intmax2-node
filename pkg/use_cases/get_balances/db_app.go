package get_balances

import (
	"context"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=get_balances_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	BackupTransfers
	BackupTransactions
	BackupDeposits
	BackupBalances
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
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
	GetBackupDeposit(conditions []string, values []interface{}) (*mDBApp.BackupDeposit, error)
	GetBackupDeposits(condition string, value interface{}) ([]*mDBApp.BackupDeposit, error)
}

type BackupBalances interface {
	CreateBackupBalance(input *backupBalance.UCPostBackupBalanceInput) (*mDBApp.BackupBalance, error)
	GetBackupBalance(conditions []string, values []interface{}) (*mDBApp.BackupBalance, error)
	GetBackupBalances(condition string, value interface{}) ([]*mDBApp.BackupBalance, error)
}
