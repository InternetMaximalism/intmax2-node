package post_backup_balance

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_balance"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct{}

func New() backup_balance.UseCasePostBackupBalance {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *backup_balance.UCPostBackupBalanceInput,
) (*backup_balance.UCPostBackupBalance, error) {
	const (
		hName                    = "UseCase PostBackupBalance"
		userKey                  = "user"
		blockNumberKey           = "block_number"
		encryptedBalanceProofKey = "encrypted_balance_proof"
		encryptedBalanceDataKey  = "encrypted_balance_data"
		encryptedTxsKey          = "encrypted_txs"
		encryptedTransfersKey    = "encrypted_transfers"
		encryptedDepositsKey     = "encrypted_deposits"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCPostBackupBalanceInputEmpty)
		return nil, ErrUCPostBackupBalanceInputEmpty
	}

	span.SetAttributes(
		attribute.String(userKey, input.DecodeUser.ToAddress().String()),
		attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
		attribute.String(encryptedBalanceProofKey, input.EncryptedBalanceProof),
		attribute.String(encryptedBalanceDataKey, input.EncryptedBalanceData),
		attribute.StringSlice(encryptedTxsKey, input.EncryptedTxs),
		attribute.StringSlice(encryptedTransfersKey, input.EncryptedTransfers),
		attribute.StringSlice(encryptedDepositsKey, input.EncryptedDeposits),
	)

	// TODO: Implement backup balance post logic here.

	resp := backup_balance.UCPostBackupBalance{
		Message: "Backup successful.",
	}

	return &resp, nil
}
