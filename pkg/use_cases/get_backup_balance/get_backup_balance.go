package get_backup_balance

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/use_cases/backup_balance"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct{}

func New() backup_balance.UseCaseGetBackupBalance {
	return &uc{}
}

func (u *uc) Do(
	ctx context.Context, input *backup_balance.UCGetBackupBalanceInput,
) (*backup_balance.UCGetBackupBalance, error) {
	const (
		hName          = "UseCase GetBackupBalance"
		userKey        = "user"
		blockNumberKey = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetBackupBalanceInputEmpty)
		return nil, ErrUCGetBackupBalanceInputEmpty
	}

	span.SetAttributes(
		attribute.String(userKey, input.DecodeUser.ToAddress().String()),
		attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
	)

	// TODO: Implement backup balance get logic here.
	encryptedBalanceProof := ""
	encryptedBalanceData := ""
	encryptedTxs := []string{}
	encryptedTransfers := []string{}
	encryptedDeposits := []string{}

	resp := backup_balance.UCGetBackupBalance{
		EncryptedBalanceProof: encryptedBalanceProof,
		EncryptedBalanceData:  encryptedBalanceData,
		EncryptedTxs:          encryptedTxs,
		EncryptedTransfers:    encryptedTransfers,
		EncryptedDeposits:     encryptedDeposits,
	}

	return &resp, nil
}
