package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	getBackupDepositByHash "intmax2-node/internal/use_cases/get_backup_deposit_by_hash"
	backupDeposit "intmax2-node/internal/use_cases/get_backup_deposits"
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
		return nil, fmt.Errorf("failed to get backup deposits from db: %w", err)
	}
	return deposits, nil
}

func GetBackupDepositByHash(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *getBackupDepositByHash.UCGetBackupDepositByHashInput,
) (*mDBApp.BackupDeposit, error) {
	deposit, err := db.GetBackupDepositByRecipientAndDepositDoubleHash(input.Recipient, input.DepositHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup deposit by hash from db: %w", err)
	}
	return deposit, nil
}
