package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
)

func GetBackupDeposit(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupDeposit.UCGetBackupDepositInput,
) error {
	fmt.Println("GetBackupDeposit: ", input)
	return nil
}
