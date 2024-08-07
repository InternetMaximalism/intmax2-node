package store_vault_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
	verifyDepositConfirmation "intmax2-node/internal/use_cases/verify_deposit_confirmation"
	ucGetBackupBalance "intmax2-node/pkg/use_cases/get_backup_balance"
	ucGetBackupDeposit "intmax2-node/pkg/use_cases/get_backup_deposit"
	ucGetBackupTransaction "intmax2-node/pkg/use_cases/get_backup_transaction"
	ucGetBackupTransfer "intmax2-node/pkg/use_cases/get_backup_transfer"
	ucGetBalances "intmax2-node/pkg/use_cases/get_balances"
	ucVerifyDepositConfirmation "intmax2-node/pkg/use_cases/get_verify_deposit_confirmation"
	ucPostBackupBalance "intmax2-node/pkg/use_cases/post_backup_balance"
	ucPostBackupDeposit "intmax2-node/pkg/use_cases/post_backup_deposit"
	ucPostBackupTransaction "intmax2-node/pkg/use_cases/post_backup_transaction"
	ucPostBackupTransfer "intmax2-node/pkg/use_cases/post_backup_transfer"
)

//go:generate mockgen -destination=mock_commands_test.go -package=store_vault_server_test -source=commands.go

type Commands interface {
	PostBackupTransfer(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCasePostBackupTransfer
	PostBackupTransaction(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCasePostBackupTransaction
	PostBackupDeposit(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCasePostBackupDeposit
	PostBackupBalance(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCasePostBackupBalance
	GetBackupTransfer(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCaseGetBackupTransfer
	GetBackupTransaction(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCaseGetBackupTransaction
	GetBackupDeposit(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCaseGetBackupDeposit
	GetBackupBalance(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBackupBalance
	GetBalances(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBalances
	GetVerifyDepositConfirmation(cfg *configs.Config, log logger.Logger, sb ServiceBlockchain) verifyDepositConfirmation.UseCaseGetVerifyDepositConfirmation
}

type commands struct{}

func NewCommands() Commands {
	return &commands{}
}

func (c *commands) PostBackupTransfer(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCasePostBackupTransfer {
	return ucPostBackupTransfer.New(cfg, log, db)
}

func (c *commands) PostBackupTransaction(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCasePostBackupTransaction {
	return ucPostBackupTransaction.New(cfg, log, db)
}

func (c *commands) PostBackupDeposit(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCasePostBackupDeposit {
	return ucPostBackupDeposit.New(cfg, log, db)
}

func (c *commands) PostBackupBalance(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCasePostBackupBalance {
	return ucPostBackupBalance.New(cfg, log, db)
}

func (c *commands) GetBalances(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBalances {
	return ucGetBalances.New(cfg, log, db)
}

func (c *commands) GetBackupTransfer(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCaseGetBackupTransfer {
	return ucGetBackupTransfer.New(cfg, log, db)
}

func (c *commands) GetBackupTransaction(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCaseGetBackupTransaction {
	return ucGetBackupTransaction.New(cfg, log, db)
}

func (c *commands) GetBackupDeposit(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCaseGetBackupDeposit {
	return ucGetBackupDeposit.New(cfg, log, db)
}

func (c *commands) GetBackupBalance(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBackupBalance {
	return ucGetBackupBalance.New(cfg, log, db)
}

func (c *commands) GetVerifyDepositConfirmation(cfg *configs.Config, log logger.Logger, sb ServiceBlockchain) verifyDepositConfirmation.UseCaseGetVerifyDepositConfirmation {
	return ucVerifyDepositConfirmation.New(cfg, log, sb)
}
