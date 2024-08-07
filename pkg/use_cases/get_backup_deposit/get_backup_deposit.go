package get_backup_deposit

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_deposit"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_deposit.UseCaseGetBackupDeposit {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_deposit.UCGetBackupDepositInput,
) (*node.GetBackupDepositResponse_Data, error) {
	const (
		hName           = "UseCase GetBackupDeposit"
		recipientKey    = "recipient"
		startBackupTime = "start_backup_time"
		limitKey        = "limit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupDepositInputEmpty)
		return nil, ErrUCGetBackupDepositInputEmpty
	}

	span.SetAttributes(
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	service.GetBackupDeposit(ctx, u.cfg, u.log, u.db, input)

	data := node.GetBackupDepositResponse_Data{
		Transactions: genTransaction(),
		Meta: &node.GetBackupDepositResponse_Meta{
			StartBlockNumber: 0,
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func genTransaction() []*node.GetBackupDepositResponse_Transaction {
	result := make([]*node.GetBackupDepositResponse_Transaction, 1)
	result[0] = &node.GetBackupDepositResponse_Transaction{
		EncryptedTx: "0x123",
		BlockNumber: 1000,
	}
	return result
}
