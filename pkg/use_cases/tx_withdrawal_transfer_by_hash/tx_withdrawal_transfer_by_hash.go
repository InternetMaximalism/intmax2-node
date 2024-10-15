package tx_withdrawal_transfer_by_hash

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_withdrawal_service"
	txWithdrawalTransferByHash "intmax2-node/internal/use_cases/tx_withdrawal_transfer_by_hash"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
) txWithdrawalTransferByHash.UseCaseTxWithdrawalTransferByHash {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(
	ctx context.Context,
	args []string,
	transferHash, userEthPrivateKey string,
) (json.RawMessage, error) {
	const (
		hName           = "UseCase TxWithdrawalTransferByHash"
		recipientKey    = "recipient"
		transferHashKey = "transfer_hash"
		emptyKey        = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(transferHashKey, transferHash),
		))
	defer span.End()

	userEthPrivateKey = strings.TrimSpace(userEthPrivateKey)
	if userEthPrivateKey == emptyKey {
		open_telemetry.MarkSpanError(spanCtx, ErrEmptyUserPrivateKey)
		return nil, ErrEmptyUserPrivateKey
	}

	transferHash = strings.TrimSpace(transferHash)
	if transferHash == emptyKey {
		open_telemetry.MarkSpanError(spanCtx, ErrEmptyTxHash)
		return nil, ErrEmptyTxHash
	}

	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, fmt.Errorf("fail to parse user private key: %v", err)
	}

	span.SetAttributes(
		attribute.String(recipientKey, wallet.WalletAddress.String()),
	)

	return service.TransferByHash(spanCtx, u.cfg, transferHash, userEthPrivateKey)
}
