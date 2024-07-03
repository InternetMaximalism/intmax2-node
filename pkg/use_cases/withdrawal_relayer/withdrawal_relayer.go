package withdrawal

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucWithdrawalRelayer "intmax2-node/internal/use_cases/withdrawal_relayer"
	service "intmax2-node/internal/withdrawal_service"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	sb  ServiceBlockchain
}

func New(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) ucWithdrawalRelayer.UseCaseWithdrawalRelayer {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context) (err error) {
	const (
		hName = "UseCase Withdrawal Relayer"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			const msg = "exec of withdrawal relayer error occurred: %w"
			err = fmt.Errorf(msg, fmt.Errorf("%+v", r))
			open_telemetry.MarkSpanError(spanCtx, err)
		}
	}()

	service.WithdrawalRelayer(spanCtx, u.cfg, u.log, u.sb)

	return err
}
