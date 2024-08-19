package store_vault_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
	getVersion "intmax2-node/internal/use_cases/get_version"
	verifyDepositConfirmation "intmax2-node/internal/use_cases/verify_deposit_confirmation"
	ucGetBackupBalances "intmax2-node/pkg/use_cases/get_backup_balances"
	ucGetBackupDeposits "intmax2-node/pkg/use_cases/get_backup_deposits"
	ucGetBackupTransactions "intmax2-node/pkg/use_cases/get_backup_transactions"
	ucGetBackupTransfers "intmax2-node/pkg/use_cases/get_backup_transfers"
	ucGetBalances "intmax2-node/pkg/use_cases/get_balances"
	ucVerifyDepositConfirmation "intmax2-node/pkg/use_cases/get_verify_deposit_confirmation"
	ucGetVersion "intmax2-node/pkg/use_cases/get_version"
	ucPostBackupBalance "intmax2-node/pkg/use_cases/post_backup_balance"
	ucPostBackupDeposit "intmax2-node/pkg/use_cases/post_backup_deposit"
	ucPostBackupTransaction "intmax2-node/pkg/use_cases/post_backup_transaction"
	ucPostBackupTransfer "intmax2-node/pkg/use_cases/post_backup_transfer"
)

//go:generate mockgen -destination=mock_commands_test.go -package=store_vault_server_test -source=commands.go

type Commands interface {
	GetVersion(version, buildTime string) getVersion.UseCaseGetVersion
	PostBackupTransfer(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCasePostBackupTransfer
	PostBackupTransaction(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCasePostBackupTransaction
	PostBackupDeposit(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCasePostBackupDeposit
	PostBackupBalance(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCasePostBackupBalance
	GetBackupTransfers(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCaseGetBackupTransfers
	GetBackupTransactions(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCaseGetBackupTransactions
	GetBackupDeposits(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCaseGetBackupDeposits
	GetBackupBalances(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBackupBalances
	GetBalances(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBalances
	GetVerifyDepositConfirmation(cfg *configs.Config, log logger.Logger, sb ServiceBlockchain) verifyDepositConfirmation.UseCaseGetVerifyDepositConfirmation
}

type commands struct{}

func NewCommands() Commands {
	return &commands{}
}

func (c *commands) GetVersion(version, buildTime string) getVersion.UseCaseGetVersion {
	return ucGetVersion.New(version, buildTime)
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

func (c *commands) GetBackupTransfers(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransfer.UseCaseGetBackupTransfers {
	return ucGetBackupTransfers.New(cfg, log, db)
}

func (c *commands) GetBackupTransactions(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupTransaction.UseCaseGetBackupTransactions {
	return ucGetBackupTransactions.New(cfg, log, db)
}

func (c *commands) GetBackupDeposits(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupDeposit.UseCaseGetBackupDeposits {
	return ucGetBackupDeposits.New(cfg, log, db)
}

func (c *commands) GetBackupBalances(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBackupBalances {
	return ucGetBackupBalances.New(cfg, log, db)
}

func (c *commands) GetBalances(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCaseGetBalances {
	return ucGetBalances.New(cfg, log, db)
}

func (c *commands) GetVerifyDepositConfirmation(cfg *configs.Config, log logger.Logger, sb ServiceBlockchain) verifyDepositConfirmation.UseCaseGetVerifyDepositConfirmation {
	return ucVerifyDepositConfirmation.New(cfg, log, sb)
}
