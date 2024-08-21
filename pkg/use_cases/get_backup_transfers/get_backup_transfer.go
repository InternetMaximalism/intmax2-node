package get_backup_transfers

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	service "intmax2-node/internal/store_vault_service"
	getBackupTransfers "intmax2-node/internal/use_cases/get_backup_transfers"
	"intmax2-node/pkg/sql_db/db_app/models"
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
) getBackupTransfers.UseCaseGetBackupTransfers {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *getBackupTransfers.UCGetBackupTransfersInput,
) (*node.GetBackupTransfersResponse_Data, error) {
	const (
		hName               = "UseCase GetBackupTransfers"
		senderKey           = "sender"
		startBlockNumberKey = "start_block_number"
		limitKey            = "limit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupTransfersInputEmpty)
		return nil, ErrUCGetBackupTransfersInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderKey, input.Sender),
		attribute.Int64(startBlockNumberKey, int64(input.StartBlockNumber)),
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	transfers, err := service.GetBackupTransfers(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return nil, err
	}

	data := node.GetBackupTransfersResponse_Data{
		Transfers: generateBackupTransfers(transfers),
		Meta: &node.GetBackupTransfersResponse_Meta{
			StartBlockNumber: input.StartBlockNumber,
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func generateBackupTransfers(transfers []*models.BackupTransfer) []*node.GetBackupTransfersResponse_Transfer {
	results := make([]*node.GetBackupTransfersResponse_Transfer, 0, len(transfers))
	for key := range transfers {
		backupTransfer := &node.GetBackupTransfersResponse_Transfer{
			Id:                transfers[key].ID,
			Recipient:         transfers[key].Recipient,
			BlockNumber:       transfers[key].BlockNumber,
			EncryptedTransfer: transfers[key].EncryptedTransfer,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: transfers[key].CreatedAt.Unix(),
				Nanos:   int32(transfers[key].CreatedAt.Nanosecond()),
			},
		}
		results = append(results, backupTransfer)
	}
	return results
}
