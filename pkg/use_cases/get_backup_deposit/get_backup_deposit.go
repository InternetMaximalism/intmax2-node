package get_backup_deposit

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_deposit"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct{}

func New() backup_deposit.UseCaseGetBackupDeposit {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *backup_deposit.UCGetBackupDepositInput,
) (*backup_deposit.UCGetBackupDeposit, error) {
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
		attribute.String(recipientKey, input.DecodeRecipient.ToAddress().String()),
		attribute.Int64(startBackupTime, int64(input.StartBackupTime)),
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	// TODO: Implement backup balance get logic here.
	deposits := make([]backup_deposit.UCGetBackupDepositContent, 0)
	meta := backup_deposit.UCGetBackupDepositMeta{
		StartBackupTime: time.Now(),
		EndBackupTime:   time.Now(),
	}
	resp := backup_deposit.UCGetBackupDeposit{
		Deposits: deposits,
		Meta:     meta,
	}

	return &resp, nil
}
