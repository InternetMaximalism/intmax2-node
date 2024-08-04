package messenger_withdrawal_relayer_mock

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	service "intmax2-node/internal/messenger"
	"intmax2-node/internal/open_telemetry"
	ucWithdrawalRelayerMock "intmax2-node/internal/use_cases/messenger_withdrawal_relayer_mock"
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
) ucWithdrawalRelayerMock.UseCaseMessengerWithdrawalRelayerMock {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context) (err error) {
	const (
		hName = "UseCase MessengerWithdrawalRelayerMock"
	)

	u.log.Infof("Starting MessengerWithdrawalRelayerMock")

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			const msg = "exec of messenger withdrawal relayer mock error occurred: %w"
			err = fmt.Errorf(msg, fmt.Errorf("%+v", r))
			open_telemetry.MarkSpanError(spanCtx, err)
		} else {
			u.log.Infof("Completed MessengerWithdrawalRelayerMock")
		}
	}()

	service.MessengerWithdrawalRelayerMock(spanCtx, u.cfg, u.log, u.db, u.sb)

	return err
}
