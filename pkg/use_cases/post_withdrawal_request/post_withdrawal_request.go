package post_withdrawal_request

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"

	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	service "intmax2-node/internal/withdrawal_service"

	"go.opentelemetry.io/otel/attribute"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
	sb  service.ServiceBlockchain
}

func New(cfg *configs.Config, log logger.Logger, db SQLDriverApp, sb service.ServiceBlockchain) postWithdrwalRequest.UseCasePostWithdrawalRequest {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context, input *postWithdrwalRequest.UCPostWithdrawalRequestInput) error {
	const (
		hName     = "UseCase PostWithdrawalRequest"
		txHashKey = "tx_hash"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCInputEmpty)
		return ErrUCInputEmpty
	}

	span.SetAttributes(
		attribute.String(txHashKey, input.TransferHash),
	)

	err := service.PostWithdrawalRequest(ctx, u.cfg, u.log, u.db, u.sb, input)
	if err != nil {
		if errors.Is(err, service.ErrWithdrawalRequestAlreadyExists) {
			return service.ErrWithdrawalRequestAlreadyExists
		}

		return fmt.Errorf("failed to post withdrawal request: %w", err)
	}
	return nil
}
