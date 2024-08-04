package transaction

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	txDeposit "intmax2-node/internal/use_cases/tx_deposit"
	txTransfer "intmax2-node/internal/use_cases/tx_transfer"
	txWithdrawal "intmax2-node/internal/use_cases/tx_withdrawal"
	ucTxDeposit "intmax2-node/pkg/use_cases/tx_deposit"
	ucTxTransfer "intmax2-node/pkg/use_cases/tx_transfer"
	ucTxWithdrawal "intmax2-node/pkg/use_cases/tx_withdrawal"
)

type Commands interface {
	SendTransferTransaction(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txTransfer.UseCaseTxTransfer
	SendDepositTransaction(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txDeposit.UseCaseTxDeposit
	SendWithdrawalTransaction(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txWithdrawal.UseCaseTxWithdrawal
}

type commands struct{}

func newCommands() Commands {
	return &commands{}
}

func (c *commands) SendTransferTransaction(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txTransfer.UseCaseTxTransfer {
	return ucTxTransfer.New(cfg, log, sb)
}

func (c *commands) SendDepositTransaction(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txDeposit.UseCaseTxDeposit {
	return ucTxDeposit.New(cfg, log, sb)
}

func (c *commands) SendWithdrawalTransaction(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txWithdrawal.UseCaseTxWithdrawal {
	return ucTxWithdrawal.New(cfg, log, sb)
}
