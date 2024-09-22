package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	models "intmax2-node/pkg/sql_db/db_app/models"
)

func PostBackupBalance(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupBalance.UCPostBackupBalanceInput,
) (*models.BackupBalance, error) {
	newBackupBalance, err := db.CreateBackupBalance(
		input.User, input.EncryptedBalanceProof, input.EncryptedBalanceData, input.Signature,
		input.EncryptedTxs, input.EncryptedTransfers, input.EncryptedDeposits,
		int64(input.BlockNumber),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup balance to db: %w", err)
	}

	return newBackupBalance, nil
}
