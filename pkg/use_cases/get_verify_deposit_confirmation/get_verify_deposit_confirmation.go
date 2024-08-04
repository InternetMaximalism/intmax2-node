package get_verify_deposit_confirmation

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/store_vault_service"
	verifyDepositConfirmation "intmax2-node/internal/use_cases/verify_deposit_confirmation"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	sb  ServiceBlockchain
}

func New(cfg *configs.Config, log logger.Logger, sb ServiceBlockchain) verifyDepositConfirmation.UseCaseGetVerifyDepositConfirmation {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(
	ctx context.Context, input *verifyDepositConfirmation.UCGetVerifyDepositConfirmationInput,
) (bool, error) {
	const (
		hName     = "UseCase GetBalances"
		depositId = "depositId"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCGetVerifyDepositConfirmationInputInputEmpty)
		return false, ErrUCGetVerifyDepositConfirmationInputInputEmpty
	}

	span.SetAttributes(
		attribute.String(depositId, input.DepositId),
	)

	confirmed, err := service.GetVerifyDepositConfirmation(ctx, u.cfg, u.log, u.sb, input)
	if err != nil {
		return false, fmt.Errorf("failed to get verify deposit confirmation: %w", err)
	}

	return confirmed, nil
}
