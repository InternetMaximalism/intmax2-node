package get_backup_deposits

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_deposit"
	"intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_deposit.UseCaseGetBackupDeposits {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_deposit.UCGetBackupDepositsInput,
) (*node.GetBackupDepositsResponse_Data, error) {
	const (
		hName           = "UseCase GetBackupDeposits"
		recipientKey    = "recipient"
		startBackupTime = "start_backup_time"
		limitKey        = "limit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupDepositsInputEmpty)
		return nil, ErrUCGetBackupDepositsInputEmpty
	}

	span.SetAttributes(
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	deposits, err := service.GetBackupDeposits(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return nil, err
	}

	data := node.GetBackupDepositsResponse_Data{
		Deposits: generateBackupDeposits(deposits),
		Meta: &node.GetBackupDepositsResponse_Meta{
			StartBlockNumber: 0,
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func generateBackupDeposits(deposits []*models.BackupDeposit) []*node.GetBackupDepositsResponse_Deposit {
	results := make([]*node.GetBackupDepositsResponse_Deposit, 0, len(deposits))
	for _, deposit := range deposits {
		backupDeposit := &node.GetBackupDepositsResponse_Deposit{
			Id:               deposit.ID,
			Recipient:        deposit.Recipient,
			BlockNumber:      uint64(deposit.BlockNumber),
			EncryptedDeposit: deposit.EncryptedDeposit,
			CreatedAt:        deposit.CreatedAt.Format(time.RFC3339),
		}
		results = append(results, backupDeposit)
	}
	return results
}
