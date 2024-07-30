package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
)

func PostBackupTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupTransaction.UCPostBackupTransactionInput,
) error {
	_, err := db.CreateBackupTransaction(input)
	if err != nil {
		return fmt.Errorf("failed to create backup transaction to db: %w", err)
	}
	return nil
}
