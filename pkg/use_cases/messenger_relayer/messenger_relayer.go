package messenger_relayer

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	service "intmax2-node/internal/messenger"
	"intmax2-node/internal/open_telemetry"
	ucMessengerRelayer "intmax2-node/internal/use_cases/messenger_relayer"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
	sb  ServiceBlockchain
}

func New(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	sb ServiceBlockchain,
) ucMessengerRelayer.UseCaseMessengerRelayer {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context) (err error) {
	const (
		hName = "UseCase MessengerRelayer"
	)

	u.log.Infof("Starting MessengerRelayer")

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			const msg = "exec of messenger relayer error occurred: %w"
			err = fmt.Errorf(msg, fmt.Errorf("%+v", r))
			open_telemetry.MarkSpanError(spanCtx, err)
		} else {
			u.log.Infof("Completed MessengerRelayer")
		}
	}()

	service.MessengerRelayer(spanCtx, u.cfg, u.log, u.db, u.sb)

	return err
}
