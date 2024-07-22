package get_backup_transaction

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_transaction"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct{}

func New() backup_transaction.UseCaseGetBackupTransaction {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *backup_transaction.UCGetBackupTransactionInput,
) (*backup_transaction.UCGetBackupTransaction, error) {
	const (
		hName               = "UseCase GetBackupTransaction"
		senderKey           = "sender"
		startBlockNumberKey = "start_block_number"
		limitKey            = "limit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupTransactionInputEmpty)
		return nil, ErrUCGetBackupTransactionInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
		attribute.Int64(startBlockNumberKey, int64(input.StartBlockNumber)),
		attribute.Int64(limitKey, int64(input.Limit)),
	)

	// TODO: Implement backup balance get logic here.
	transactions := make([]backup_transaction.UCGetBackupTransactionContent, 0)
	meta := backup_transaction.UCGetBackupTransactionMeta{
		StartBlockNumber: 0,
		EndBlockNumber:   0,
	}
	resp := backup_transaction.UCGetBackupTransaction{
		Transactions: transactions,
		Meta:         meta,
	}

	return &resp, nil
}
