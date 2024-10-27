package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupProof "intmax2-node/internal/use_cases/backup_balance_proof"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func GetBackupSenderProofs(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupProof.UCGetBackupBalanceProofsInput,
) ([]*mDBApp.BackupSenderProof, error) {
	balances, err := db.GetBackupSenderProofsByHashes(input.Hashes)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender balance proofs from db: %w", err)
	}
	return balances, nil
}
