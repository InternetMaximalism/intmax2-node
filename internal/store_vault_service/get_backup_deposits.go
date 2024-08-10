package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func GetBackupDeposits(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupDeposit.UCGetBackupDepositsInput,
) ([]*mDBApp.BackupDeposit, error) {
	deposits, err := db.GetBackupDeposits("recipient", input.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup deposit from db: %w", err)
	}
	return deposits, nil
}
