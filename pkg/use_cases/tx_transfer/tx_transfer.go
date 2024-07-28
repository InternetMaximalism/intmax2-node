package tx_transfer

import (
	"context"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_transfer_service"
	txTransfer "intmax2-node/internal/use_cases/tx_transfer"

	"go.opentelemetry.io/otel/attribute"
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
) txTransfer.UseCaseTxTransfer {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
		sb:  sb,
	}
}

func (u *uc) Do(ctx context.Context, args []string, amount, recipientAddressStr, userPrivateKey string) (err error) {
	const (
		hName     = "UseCase TxTransfer"
		senderKey = "sender"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	// The userPrivateKey is acceptable in either format:
	// it may include the '0x' prefix at the beginning,
	// or it can be provided without this prefix.
	userAccount, err := intMaxAcc.NewPrivateKeyFromString(userPrivateKey)
	if err != nil {
		return err
	}

	userAddress := userAccount.ToAddress()
	span.SetAttributes(
		attribute.String(senderKey, userAddress.String()),
	)

	service.SendTransferTransaction(spanCtx, u.cfg, u.log, u.db, u.sb, args, amount, recipientAddressStr, userPrivateKey)

	return nil
}
