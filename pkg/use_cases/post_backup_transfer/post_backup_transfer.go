package post_backup_transfer

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_transfer"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct{}

func New() backup_transfer.UseCasePostBackupTransfer {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *backup_transfer.UCPostBackupTransferInput,
) (*backup_transfer.UCPostBackupTransfer, error) {
	const (
		hName                = "UseCase PostBackupTransfer"
		recipientKey         = "recipient"
		blockNumberKey       = "block_number"
		encryptedTransferKey = "encrypted_transfer"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCPostBackupTransferInputEmpty)
		return nil, ErrUCPostBackupTransferInputEmpty
	}

	span.SetAttributes(
		attribute.String(recipientKey, input.DecodeRecipient.ToAddress().String()),
		attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
		attribute.String(encryptedTransferKey, input.EncryptedTransfer),
	)

	// TODO: Implement backup balance post logic here.

	resp := backup_transfer.UCPostBackupTransfer{
		Message: "Transfer data backup successful.",
	}

	return &resp, nil
}
