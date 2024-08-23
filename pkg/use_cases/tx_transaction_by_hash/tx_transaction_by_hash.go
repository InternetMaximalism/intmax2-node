package tx_transaction_by_hash

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_transfer_service"
	txTransactionByHash "intmax2-node/internal/use_cases/tx_transaction_by_hash"
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
) txTransactionByHash.UseCaseTxTransactionByHash {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(
	ctx context.Context,
	args []string,
	txHash, userEthPrivateKey string,
) (json.RawMessage, error) {
	const (
		hName     = "UseCase TxTransactionByHash"
		senderKey = "sender"
		txHashKey = "tx_hash"
		emptyKey  = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(txHashKey, txHash),
		))
	defer span.End()

	userEthPrivateKey = strings.TrimSpace(userEthPrivateKey)
	if userEthPrivateKey == emptyKey {
		open_telemetry.MarkSpanError(spanCtx, ErrEmptyUserPrivateKey)
		return nil, ErrEmptyUserPrivateKey
	}

	txHash = strings.TrimSpace(txHash)
	if txHash == emptyKey {
		open_telemetry.MarkSpanError(spanCtx, ErrEmptyTxHash)
		return nil, ErrEmptyTxHash
	}

	wallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(userEthPrivateKey)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, fmt.Errorf("fail to parse user private key: %v", err)
	}

	// The userPrivateKey is acceptable in either format:
	// it may include the '0x' prefix at the beginning,
	// or it can be provided without this prefix.
	userAccount, err := intMaxAcc.NewPrivateKeyFromString(wallet.IntMaxPrivateKey)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, err
	}

	userAddress := userAccount.ToAddress()

	span.SetAttributes(
		attribute.String(senderKey, userAddress.String()),
	)

	return service.TransactionByHash(spanCtx, u.cfg, txHash, userEthPrivateKey)
}
