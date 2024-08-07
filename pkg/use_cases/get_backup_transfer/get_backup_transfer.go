package get_backup_transfer

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	service "intmax2-node/internal/store_vault_service"
	"intmax2-node/internal/use_cases/backup_transfer"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) backup_transfer.UseCaseGetBackupTransfer {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *backup_transfer.UCGetBackupTransferInput,
) (*node.GetBackupTransferResponse_Data, error) {
	const (
		hName           = "UseCase GetBackupTransfer"
		recipientKey    = "recipient"
		startBackupTime = "start_backup_time"
		limitKey        = "limit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupTransferInputEmpty)
		return nil, ErrUCGetBackupTransferInputEmpty
	}

	span.SetAttributes(
		// attribute.String(recipientKey, input.DecodeRecipient.ToAddress().String()),
		// attribute.Int64(startBackupTime, int64(input.StartBackupTime)),
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	// TODO: Implement backup balance get logic here.
	service.GetBackupTransfer(ctx, u.cfg, u.log, u.db, input)

	data := node.GetBackupTransferResponse_Data{
		Transactions: genTransaction(),
		Meta: &node.GetBackupTransferResponse_Meta{
			StartBlockNumber: 0,
			EndBlockNumber:   0,
		},
	}

	return &data, nil
}

func genTransaction() []*node.GetBackupTransferResponse_Transaction {
	result := make([]*node.GetBackupTransferResponse_Transaction, 1)
	result[0] = &node.GetBackupTransferResponse_Transaction{
		EncryptedTx: "0x123",
		BlockNumber: 1000,
	}
	return result
}
