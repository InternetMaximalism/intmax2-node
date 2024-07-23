package sync_balance

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	balanceChecker "intmax2-node/internal/use_cases/balance_checker"
	ucBalanceChecker "intmax2-node/pkg/use_cases/sync_balance"
)

type Commands interface {
	SyncBalance(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
		sb ServiceBlockchain,
	) balanceChecker.UseCaseBalanceChecker
}

type commands struct{}

func newCommands() Commands {
	return &commands{}
}

func (c *commands) SyncBalance(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) balanceChecker.UseCaseBalanceChecker {
	return ucBalanceChecker.New(cfg, log, db, sb)
}
