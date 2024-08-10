package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func GetBackupTransactions(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupTransaction.UCGetBackupTransactionsInput,
) ([]*mDBApp.BackupTransaction, error) {
	transactions, err := db.GetBackupTransactions("sender", input.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup transactions from db: %w", err)
	}
	return transactions, nil
}
