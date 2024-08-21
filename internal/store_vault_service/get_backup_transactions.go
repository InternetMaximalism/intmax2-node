package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	getBackupTransactionByHash "intmax2-node/internal/use_cases/get_backup_transaction_by_hash"
	getBackupTransactions "intmax2-node/internal/use_cases/get_backup_transactions"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func GetBackupTransactions(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *getBackupTransactions.UCGetBackupTransactionsInput,
) ([]*mDBApp.BackupTransaction, error) {
	transactions, err := db.GetBackupTransactions("sender", input.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup transactions from db: %w", err)
	}
	return transactions, nil
}

func GetBackupTransactionByHash(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *getBackupTransactionByHash.UCGetBackupTransactionByHashInput,
) (*mDBApp.BackupTransaction, error) {
	transaction, err := db.GetBackupTransactionBySenderAndTxDoubleHash(input.Sender, input.TxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup transaction by hash from db: %w", err)
	}
	return transaction, nil
}
