package store_vault_service

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
)

func PostBackupBalance(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	input *backupBalance.UCPostBackupBalanceInput,
) error {
	fmt.Println("PostBackupBalance: ", input)
	return nil
}