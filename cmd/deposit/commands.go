package deposit

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	depositAnalyzer "intmax2-node/internal/use_cases/deposit_analyzer"
	depositRelayer "intmax2-node/internal/use_cases/deposit_relayer"
	ucDepositAnalyzer "intmax2-node/pkg/use_cases/deposit_analyzer"
	ucDepositRelayer "intmax2-node/pkg/use_cases/deposit_relayer"
)

//go:generate mockgen -destination=mock_command.go -package=deposit -source=commands.go

type Commands interface {
	DepositAnalyzer(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
		sb ServiceBlockchain,
	) depositAnalyzer.UseCaseDepositAnalyzer
	DepositRelayer(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
		sb ServiceBlockchain,
	) depositRelayer.UseCaseDepositRelayer
}

type commands struct{}

func newCommands() Commands {
	return &commands{}
}

func (c *commands) DepositAnalyzer(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) depositAnalyzer.UseCaseDepositAnalyzer {
	return ucDepositAnalyzer.New(cfg, log, db, sb)
}

func (c *commands) DepositRelayer(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) depositRelayer.UseCaseDepositRelayer {
	return ucDepositRelayer.New(cfg, log, db, sb)
}
