package post_backup_deposit

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/store_vault_service"
	postBackupDeposit "intmax2-node/internal/use_cases/post_backup_deposit"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp) postBackupDeposit.UseCasePostBackupDeposit {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context, input *postBackupDeposit.UCPostBackupDepositInput,
) error {
	const (
		hName               = "UseCase PostBackupDeposit"
		recipientKey        = "recipient"
		blockNumberKey      = "block_number"
		depositHashKey      = "deposit_hash"
		encryptedDepositKey = "encrypted_deposit"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCPostBackupDepositInputEmpty)
		return ErrUCPostBackupDepositInputEmpty
	}

	span.SetAttributes(
		attribute.String(recipientKey, input.Recipient),
		attribute.Int64(blockNumberKey, input.BlockNumber),
		attribute.String(depositHashKey, input.DepositHash),
		attribute.String(encryptedDepositKey, input.EncryptedDeposit),
	)

	err := service.PostBackupDeposit(ctx, u.cfg, u.log, u.db, input)
	if err != nil {
		return fmt.Errorf("failed to post backup deposit: %w", err)
	}

	return nil
}
