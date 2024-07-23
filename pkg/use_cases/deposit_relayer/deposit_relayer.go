package deposit_relayer

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	service "intmax2-node/internal/deposit_service"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	depositRelayer "intmax2-node/internal/use_cases/deposit_relayer"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
	sb  ServiceBlockchain
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) depositRelayer.UseCaseDepositRelayer {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context) (err error) {
	const (
		hName = "UseCase DepositRelayer"
	)

	u.log.Infof("Starting DepositRelayer")

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			const msg = "exec of deposit analyzer error occurred: %w"
			err = fmt.Errorf(msg, fmt.Errorf("%+v", r))
			open_telemetry.MarkSpanError(spanCtx, err)
		} else {
			u.log.Infof("Completed DepositRelayer")
		}
	}()

	service.DepositRelayer(spanCtx, u.cfg, u.log, u.db, u.sb)

	return err
}
