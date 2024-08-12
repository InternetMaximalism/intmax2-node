package get_backup_transfers

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_transfer"
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

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_transfer.UseCaseGetBackupTransfers {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_transfer.UCGetBackupTransferInput,
) (*node.GetBackupTransfersResponse_Data, error) {
	const (
		hName           = "UseCase GetBackupTransfers"
		recipientKey    = "recipient"
		startBackupTime = "start_backup_time"
		limitKey        = "limit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupTransfersInputEmpty)
		return nil, ErrUCGetBackupTransfersInputEmpty
	}

	span.SetAttributes(
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	transfers, err := service.GetBackupTransfers(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return nil, err
	}

	data := node.GetBackupTransfersResponse_Data{
		Transfers: generateBackupTransfers(transfers),
		Meta: &node.GetBackupTransfersResponse_Meta{
			StartBlockNumber: uint64(input.StartBlockNumber),
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func generateBackupTransfers(transfers []*models.BackupTransfer) []*node.GetBackupTransfersResponse_Transfer {
	results := make([]*node.GetBackupTransfersResponse_Transfer, 0, len(transfers))
	for _, transfer := range transfers {
		backupTransfer := &node.GetBackupTransfersResponse_Transfer{
			Id:                transfer.ID,
			Recipient:         transfer.Recipient,
			BlockNumber:       uint64(transfer.BlockNumber),
			EncryptedTransfer: transfer.EncryptedTransfer,
			CreatedAt:         transfer.CreatedAt.Format(time.RFC3339),
		}
		results = append(results, backupTransfer)
	}
	return results
}
