package transaction

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	balanceChecker "intmax2-node/internal/use_cases/tx_transfer"
	ucTxDeposit "intmax2-node/pkg/use_cases/tx_deposit"
	ucTxTransfer "intmax2-node/pkg/use_cases/tx_transfer"
	ucTxWithdraw "intmax2-node/pkg/use_cases/tx_withdrawal"
)

type Commands interface {
	SendTransferTransaction(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) balanceChecker.UseCaseTxTransfer
	SendDepositTransaction(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) balanceChecker.UseCaseTxTransfer
	SendWithdrawalTransaction(
		cfg *configs.Config,
		log logger.Logger,
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
	sb ServiceBlockchain,
) balanceChecker.UseCaseTxTransfer {
	return ucTxTransfer.New(cfg, log, sb)
}

func (c *commands) SendDepositTransaction(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) balanceChecker.UseCaseTxTransfer {
	return ucTxDeposit.New(cfg, log, sb)
}

func (c *commands) SendWithdrawalTransaction(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) balanceChecker.UseCaseTxTransfer {
	return ucTxWithdraw.New(cfg, log, sb)
}
