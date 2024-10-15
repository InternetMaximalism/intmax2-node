package get_backup_transfer_by_hash

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	service "intmax2-node/internal/store_vault_service"
	getBackupTransferByHash "intmax2-node/internal/use_cases/get_backup_transfer_by_hash"

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
) getBackupTransferByHash.UseCaseGetBackupTransferByHash {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *getBackupTransferByHash.UCGetBackupTransferByHashInput,
) (*node.GetBackupTransferByHashResponse_Data, error) {
	const (
		hName           = "UseCase GetBackupTransferByHash"
		recipientKey    = "recipient"
		transferHashKey = "transfer_hash"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupTransferByHashInputEmpty)
		return nil, ErrUCGetBackupTransferByHashInputEmpty
	}

	span.SetAttributes(
		attribute.String(transferHashKey, input.TransferHash),
		attribute.String(recipientKey, input.Recipient),
	)

	transfer, err := service.GetBackupTransferByHash(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return nil, err
	}

	data := node.GetBackupTransferByHashResponse_Data{
		Transfer: &node.GetBackupTransferByHashResponse_Transfer{
			Id:                transfer.ID,
			Recipient:         transfer.Recipient,
			BlockNumber:       transfer.BlockNumber,
			EncryptedTransfer: transfer.EncryptedTransfer,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: transfer.CreatedAt.Unix(),
				Nanos:   int32(transfer.CreatedAt.Nanosecond()),
			},
		},
	}

	return &data, nil
}
