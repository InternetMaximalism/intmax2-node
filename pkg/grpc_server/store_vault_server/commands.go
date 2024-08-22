package store_vault_server

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	getBackupDeposits "intmax2-node/internal/use_cases/get_backup_deposits"
	getBackupDepositsList "intmax2-node/internal/use_cases/get_backup_deposits_list"
	getBackupTransactionByHash "intmax2-node/internal/use_cases/get_backup_transaction_by_hash"
	getBackupTransactions "intmax2-node/internal/use_cases/get_backup_transactions"
	getBackupTransactionsList "intmax2-node/internal/use_cases/get_backup_transactions_list"
	getBackupTransfers "intmax2-node/internal/use_cases/get_backup_transfers"
	getVersion "intmax2-node/internal/use_cases/get_version"
	postBackupDeposit "intmax2-node/internal/use_cases/post_backup_deposit"
	postBackupTransaction "intmax2-node/internal/use_cases/post_backup_transaction"
	postBackupTransfer "intmax2-node/internal/use_cases/post_backup_transfer"
	verifyDepositConfirmation "intmax2-node/internal/use_cases/verify_deposit_confirmation"
	ucGetBackupBalances "intmax2-node/pkg/use_cases/get_backup_balances"
	ucGetBackupDeposits "intmax2-node/pkg/use_cases/get_backup_deposits"
	ucGetBackupDepositsList "intmax2-node/pkg/use_cases/get_backup_deposits_list"
	ucGetBackupTransactionByHash "intmax2-node/pkg/use_cases/get_backup_transaction_by_hash"
	ucGetBackupTransactions "intmax2-node/pkg/use_cases/get_backup_transactions"
	ucGetBackupTransactionsList "intmax2-node/pkg/use_cases/get_backup_transactions_list"
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
	PostBackupTransfer(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) postBackupTransfer.UseCasePostBackupTransfer
	PostBackupTransaction(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) postBackupTransaction.UseCasePostBackupTransaction
	PostBackupDeposit(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) postBackupDeposit.UseCasePostBackupDeposit
	PostBackupBalance(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCasePostBackupBalance
	GetBackupTransfers(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) getBackupTransfers.UseCaseGetBackupTransfers
	GetBackupTransactions(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) getBackupTransactions.UseCaseGetBackupTransactions
	GetBackupTransactionsList(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) getBackupTransactionsList.UseCaseGetBackupTransactionsList
	GetBackupTransactionByHash(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) getBackupTransactionByHash.UseCaseGetBackupTransactionByHash
	GetBackupDeposits(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) getBackupDeposits.UseCaseGetBackupDeposits
	GetBackupDepositsList(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
	) getBackupDepositsList.UseCaseGetBackupDepositsList
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

func (c *commands) PostBackupTransfer(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) postBackupTransfer.UseCasePostBackupTransfer {
	return ucPostBackupTransfer.New(cfg, log, db)
}

func (c *commands) PostBackupTransaction(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) postBackupTransaction.UseCasePostBackupTransaction {
	return ucPostBackupTransaction.New(cfg, log, db)
}

func (c *commands) PostBackupDeposit(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) postBackupDeposit.UseCasePostBackupDeposit {
	return ucPostBackupDeposit.New(cfg, log, db)
}

func (c *commands) PostBackupBalance(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backupBalance.UseCasePostBackupBalance {
	return ucPostBackupBalance.New(cfg, log, db)
}

func (c *commands) GetBackupTransfers(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupTransfers.UseCaseGetBackupTransfers {
	return ucGetBackupTransfers.New(cfg, log, db)
}

func (c *commands) GetBackupTransactions(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupTransactions.UseCaseGetBackupTransactions {
	return ucGetBackupTransactions.New(cfg, log, db)
}

func (c *commands) GetBackupTransactionsList(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupTransactionsList.UseCaseGetBackupTransactionsList {
	return ucGetBackupTransactionsList.New(cfg, log, db)
}

func (c *commands) GetBackupTransactionByHash(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupTransactionByHash.UseCaseGetBackupTransactionByHash {
	return ucGetBackupTransactionByHash.New(cfg, log, db)
}

func (c *commands) GetBackupDeposits(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupDeposits.UseCaseGetBackupDeposits {
	return ucGetBackupDeposits.New(cfg, log, db)
}

func (c *commands) GetBackupDepositsList(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupDepositsList.UseCaseGetBackupDepositsList {
	return ucGetBackupDepositsList.New(cfg, log, db)
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
