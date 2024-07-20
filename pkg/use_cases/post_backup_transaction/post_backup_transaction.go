package post_backup_transaction

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_transaction"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct{}

func New() backup_transaction.UseCasePostBackupTransaction {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *backup_transaction.UCPostBackupTransactionInput,
) (*backup_transaction.UCPostBackupTransaction, error) {
	const (
		hName          = "UseCase PostBackupTransaction"
		senderKey      = "sender"
		blockNumberKey = "block_number"
		encryptedTxKey = "encrypted_tx"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCPostBackupTransactionInputEmpty)
		return nil, ErrUCPostBackupTransactionInputEmpty
	}

	span.SetAttributes(
		attribute.String(senderKey, input.DecodeSender.ToAddress().String()),
		attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
		attribute.String(encryptedTxKey, input.EncryptedTx),
	)

	// TODO: Implement backup balance post logic here.

	resp := backup_transaction.UCPostBackupTransaction{
		Message: "Transaction data backup successful.",
	}

	return &resp, nil
}
