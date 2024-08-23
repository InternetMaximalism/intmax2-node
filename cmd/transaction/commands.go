package transaction

import (
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	txClaim "intmax2-node/internal/use_cases/tx_claim"
	txDeposit "intmax2-node/internal/use_cases/tx_deposit"
	txTransactionByHash "intmax2-node/internal/use_cases/tx_transaction_by_hash"
	txTransactionsList "intmax2-node/internal/use_cases/tx_transactions_list"
	txTransfer "intmax2-node/internal/use_cases/tx_transfer"
	txWithdrawal "intmax2-node/internal/use_cases/tx_withdrawal"
	ucTxClaim "intmax2-node/pkg/use_cases/tx_claim"
	ucTxDeposit "intmax2-node/pkg/use_cases/tx_deposit"
	ucTxTransactionByHash "intmax2-node/pkg/use_cases/tx_transaction_by_hash"
	ucTxTransactionsList "intmax2-node/pkg/use_cases/tx_transactions_list"
	ucTxTransfer "intmax2-node/pkg/use_cases/tx_transfer"
	ucTxWithdrawal "intmax2-node/pkg/use_cases/tx_withdrawal"
)

type Commands interface {
	SendTransferTransaction(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txTransfer.UseCaseTxTransfer
	SenderTransactionsList(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txTransactionsList.UseCaseTxTransactionsList
	SenderTransactionByHash(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txTransactionByHash.UseCaseTxTransactionByHash
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
	SendClaimWithdrawals(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txClaim.UseCaseTxClaim
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

func (c *commands) SenderTransactionsList(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txTransactionsList.UseCaseTxTransactionsList {
	return ucTxTransactionsList.New(cfg, log, sb)
}

func (c *commands) SenderTransactionByHash(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txTransactionByHash.UseCaseTxTransactionByHash {
	return ucTxTransactionByHash.New(cfg, log, sb)
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

func (c *commands) SendClaimWithdrawals(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txClaim.UseCaseTxClaim {
	return ucTxClaim.New(cfg, log, sb)
}
