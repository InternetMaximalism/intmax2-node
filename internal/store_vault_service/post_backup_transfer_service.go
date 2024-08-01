package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
)

func PostBackupTransfer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupTransfer.UCPostBackupTransferInput,
) error {
	_, err := db.CreateBackupTransfer(input)
	if err != nil {
		return fmt.Errorf("failed to create backup transfer to db: %w", err)
	}
	return nil
}
