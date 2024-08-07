package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
)

func GetBackupTransfer(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupTransfer.UCGetBackupTransferInput,
) error {
	fmt.Println("GetBackupTransfer: ", input)
	return nil
}
