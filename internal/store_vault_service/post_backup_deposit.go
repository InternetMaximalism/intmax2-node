package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	postBackupDeposit "intmax2-node/internal/use_cases/post_backup_deposit"
)

func PostBackupDeposit(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *postBackupDeposit.UCPostBackupDepositInput,
) error {
	_, err := db.CreateBackupDeposit(
		input.Recipient, input.DepositHash, input.EncryptedDeposit, input.BlockNumber,
	)
	if err != nil {
		return fmt.Errorf("failed to create backup deposit to db: %w", err)
	}
	return nil
}
