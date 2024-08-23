package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	postBackupTransaction "intmax2-node/internal/use_cases/post_backup_transaction"
)

func PostBackupTransaction(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *postBackupTransaction.UCPostBackupTransactionInput,
) error {
	_, err := db.CreateBackupTransaction(
		input.Sender, input.TxHash, input.EncryptedTx, input.Signature, int64(input.BlockNumber),
	)
	if err != nil {
		return fmt.Errorf("failed to create backup transaction to db: %w", err)
	}
	return nil
}
