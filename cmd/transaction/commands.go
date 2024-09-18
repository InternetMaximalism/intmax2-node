package transaction

import (
	"intmax2-node/configs"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	txClaim "intmax2-node/internal/use_cases/tx_claim"
	txDeposit "intmax2-node/internal/use_cases/tx_deposit"
	txDepositByHashIncoming "intmax2-node/internal/use_cases/tx_deposit_by_hash_incoming"
	txDepositListIncoming "intmax2-node/internal/use_cases/tx_deposits_list_incoming"
	txTransactionByHash "intmax2-node/internal/use_cases/tx_transaction_by_hash"
	txTransactionsList "intmax2-node/internal/use_cases/tx_transactions_list"
	txTransfer "intmax2-node/internal/use_cases/tx_transfer"
	txWithdrawal "intmax2-node/internal/use_cases/tx_withdrawal"
	ucTxClaim "intmax2-node/pkg/use_cases/tx_claim"
	ucTxDeposit "intmax2-node/pkg/use_cases/tx_deposit"
	ucTxDepositByHashIncoming "intmax2-node/pkg/use_cases/tx_deposit_by_hash_incoming"
	ucTxDepositListIncoming "intmax2-node/pkg/use_cases/tx_deposits_list_incoming"
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
		dbApp block_validity_prover.SQLDriverApp,
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
	ReceiverDepositsListIncoming(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txDepositListIncoming.UseCaseTxDepositsListIncoming
	ReceiverDepositByHashIncoming(
		cfg *configs.Config,
		log logger.Logger,
		sb ServiceBlockchain,
	) txDepositByHashIncoming.UseCaseTxDepositByHashIncoming
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
	db block_validity_prover.SQLDriverApp,
) txTransfer.UseCaseTxTransfer {
	return ucTxTransfer.New(cfg, log, sb, db)
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

func (c *commands) ReceiverDepositsListIncoming(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txDepositListIncoming.UseCaseTxDepositsListIncoming {
	return ucTxDepositListIncoming.New(cfg, log, sb)
}

func (c *commands) ReceiverDepositByHashIncoming(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) txDepositByHashIncoming.UseCaseTxDepositByHashIncoming {
	return ucTxDepositByHashIncoming.New(cfg, log, sb)
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
