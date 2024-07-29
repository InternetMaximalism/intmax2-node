package store_vault_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
	ucBalances "intmax2-node/pkg/use_cases/get_balances"
	ucBackupDeposit "intmax2-node/pkg/use_cases/post_backup_deposit"
	ucBackupTransaction "intmax2-node/pkg/use_cases/post_backup_transaction"
	ucBackupTransfer "intmax2-node/pkg/use_cases/post_backup_transfer"
)

//go:generate mockgen -destination=mock_commands_test.go -package=store_vault_server_test -source=commands.go

type Commands interface {
	PostBackupTransfer(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCasePostBackupTransfer
	PostBackupTransaction(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCasePostBackupTransaction
	PostBackupDeposit(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCasePostBackupDeposit
	GetBalances(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBalances
}

type commands struct{}

func NewCommands() Commands {
	return &commands{}
}

func (c *commands) PostBackupTransfer(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCasePostBackupTransfer {
	return ucBackupTransfer.New(cfg, log, db)
}

func (c *commands) PostBackupTransaction(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCasePostBackupTransaction {
	return ucBackupTransaction.New(cfg, log, db)
}

func (c *commands) PostBackupDeposit(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCasePostBackupDeposit {
	return ucBackupDeposit.New(cfg, log, db)
}

func (c *commands) GetBalances(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBalances {
	return ucBalances.New(cfg, log, db)
}
