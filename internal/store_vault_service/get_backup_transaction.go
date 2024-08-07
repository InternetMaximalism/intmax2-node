package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
)

func GetBackupTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupTransaction.UCGetBackupTransactionInput,
) error {
	fmt.Println("GetBackupTransaction: ", input)
	return nil
}
