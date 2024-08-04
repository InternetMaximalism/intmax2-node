package tx_resume_withdrawal

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_withdrawal_service"
	txResumeWithdrawal "intmax2-node/internal/use_cases/tx_resume_withdrawal"
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
) txResumeWithdrawal.UseCaseTxResumeWithdrawal {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context, recipientAddressHex string) (err error) {
	const (
		hName     = "UseCase TxTransfer"
		senderKey = "sender"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	service.ResumeWithdrawalRequest(spanCtx, u.cfg, u.log, recipientAddressHex)

	return nil
}
