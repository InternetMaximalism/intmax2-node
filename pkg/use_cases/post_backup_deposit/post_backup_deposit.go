package post_backup_deposit

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_deposit"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct{}

func New() backup_deposit.UseCasePostBackupDeposit {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *backup_deposit.UCPostBackupDepositInput,
) (*backup_deposit.UCPostBackupDeposit, error) {
	const (
		hName               = "UseCase PostBackupDeposit"
		recipientKey        = "recipient"
		blockNumberKey      = "block_number"
		encryptedDepositKey = "encrypted_deposit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCPostBackupDepositInputEmpty)
		return nil, ErrUCPostBackupDepositInputEmpty
	}

	span.SetAttributes(
		attribute.String(recipientKey, input.DecodeRecipient.ToAddress().String()),
		attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
		attribute.String(encryptedDepositKey, input.EncryptedDeposit),
	)

	// TODO: Implement backup balance post logic here.

	resp := backup_deposit.UCPostBackupDeposit{
		Message: "Transfer data backup successful.",
	}

	return &resp, nil
}
