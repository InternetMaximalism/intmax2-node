package get_backup_balances

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_balance"
	"intmax2-node/pkg/sql_db/db_app/models"
	"time"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_balance.UseCaseGetBackupBalances {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_balance.UCGetBackupBalancesInput,
) (*node.GetBackupBalancesResponse_Data, error) {
	const (
		hName          = "UseCase GetBackupBalances"
		userKey        = "user"
		blockNumberKey = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupBalancesInputEmpty)
		return nil, ErrUCGetBackupBalancesInputEmpty
	}

	balances, err := service.GetBackupBalances(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return nil, err
	}

	data := node.GetBackupBalancesResponse_Data{
		Balances: generateBackupBalances(balances),
		Meta: &node.GetBackupBalancesResponse_Meta{
			StartBlockNumber: uint64(input.StartBlockNumber),
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func generateBackupBalances(balances []*models.BackupBalance) []*node.GetBackupBalancesResponse_Balance {
	results := make([]*node.GetBackupBalancesResponse_Balance, 0, len(balances))
	for _, balance := range balances {
		backupBalance := &node.GetBackupBalancesResponse_Balance{
			Id:                    balance.ID,
			UserAddress:           balance.UserAddress,
			EncryptedBalanceProof: balance.EncryptedBalanceProof,
			EncryptedBalanceData:  balance.EncryptedBalanceData,
			EncryptedTxs:          balance.EncryptedTxs,
			EncryptedTransfers:    balance.EncryptedTransfers,
			EncryptedDeposits:     balance.EncryptedDeposits,
			BlockNumber:           uint64(balance.BlockNumber),
			Signature:             balance.Signature,
			CreatedAt:             balance.CreatedAt.Format(time.RFC3339),
		}
		results = append(results, backupBalance)
	}
	return results
}
