package get_backup_deposits

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	service "intmax2-node/internal/store_vault_service"
	getBackupDeposits "intmax2-node/internal/use_cases/get_backup_deposits"
	"intmax2-node/pkg/sql_db/db_app/models"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) getBackupDeposits.UseCaseGetBackupDeposits {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *getBackupDeposits.UCGetBackupDepositsInput,
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
			StartBlockNumber: input.StartBlockNumber,
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
			CreatedAt: &timestamppb.Timestamp{
				Seconds: deposit.CreatedAt.Unix(),
				Nanos:   int32(deposit.CreatedAt.Nanosecond()),
			},
		}
		results = append(results, backupDeposit)
	}
	return results
}
