package sync_balance

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	balanceChecker "intmax2-node/internal/use_cases/balance_checker"
	ucGetBalance "intmax2-node/pkg/use_cases/get_balance"
)

type Commands interface {
	GetBalance(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) balanceChecker.UseCaseBalanceChecker
}

type commands struct{}

func newCommands() Commands {
	return &commands{}
}

func (c *commands) GetBalance(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) balanceChecker.UseCaseBalanceChecker {
	return ucGetBalance.New(cfg, log, sb)
}
