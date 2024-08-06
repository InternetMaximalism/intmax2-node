package tx_claim

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_claim_service"
	txClaim "intmax2-node/internal/use_cases/tx_claim"
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
) txClaim.UseCaseTxClaim {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context, args []string, recipientEthPrivateKey string) (err error) {
	const (
		hName     = "UseCase TxTransfer"
		senderKey = "sender"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	service.ClaimWithdrawals(spanCtx, u.cfg, u.log, u.sb, recipientEthPrivateKey)

	return nil
}
