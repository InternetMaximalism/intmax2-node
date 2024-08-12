package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func GetBackupBalances(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupBalance.UCGetBackupBalancesInput,
) ([]*mDBApp.BackupBalance, error) {
	balances, err := db.GetBackupBalances("user_address", input.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup balances from db: %w", err)
	}
	return balances, nil
}
