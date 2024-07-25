package transaction

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	balanceChecker "intmax2-node/internal/use_cases/tx_transfer"
	ucTxTransfer "intmax2-node/pkg/use_cases/tx_transfer"
)

type Commands interface {
	SendTransferTransaction(
		cfg *configs.Config,
		log logger.Logger,
		db SQLDriverApp,
		sb ServiceBlockchain,
	) balanceChecker.UseCaseTxTransfer
}

type commands struct{}

func newCommands() Commands {
	return &commands{}
}

func (c *commands) SendTransferTransaction(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) balanceChecker.UseCaseTxTransfer {
	return ucTxTransfer.New(cfg, log, db, sb)
}
