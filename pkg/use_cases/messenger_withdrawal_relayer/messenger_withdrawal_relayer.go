package messenger_withdrawal_relayer

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	service "intmax2-node/internal/messenger"
	"intmax2-node/internal/open_telemetry"
	ucWithdrawalRelayer "intmax2-node/internal/use_cases/messenger_withdrawal_relayer"
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
) ucWithdrawalRelayer.UseCaseMessengerWithdrawalRelayer {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context) (err error) {
	const (
		hName = "UseCase MessengerWithdrawalRelayer"
	)

	u.log.Infof("Starting MessengerWithdrawalRelayer")

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			const msg = "exec of messenger withdrawal relayer error occurred: %w"
			err = fmt.Errorf(msg, fmt.Errorf("%+v", r))
			open_telemetry.MarkSpanError(spanCtx, err)
		} else {
			u.log.Infof("Completed MessengerWithdrawalRelayer")
		}
	}()

	service.MessengerWithdrawalRelayer(spanCtx, u.cfg, u.log, u.sb)

	return err
}
