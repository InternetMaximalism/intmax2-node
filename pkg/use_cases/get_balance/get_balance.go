package get_balance

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	service "intmax2-node/internal/balance_service"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	balanceChecker "intmax2-node/internal/use_cases/balance_checker"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	sb  ServiceBlockchain
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	sb ServiceBlockchain,
) balanceChecker.UseCaseBalanceChecker {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context, args []string, userEthPrivateKey string) (err error) {
	const (
		hName = "UseCase GetBalance"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			const msg = "exec of fetching balance error occurred: %w"
			err = fmt.Errorf(msg, fmt.Errorf("%+v", r))
			open_telemetry.MarkSpanError(spanCtx, err)
		}
	}()

	err = service.GetBalance(spanCtx, u.cfg, u.log, u.sb, args, userEthPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	return err
}
