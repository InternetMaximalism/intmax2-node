package tx_transactions_list

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/open_telemetry"
	service "intmax2-node/internal/tx_transfer_service"
	txTransfersList "intmax2-node/internal/use_cases/tx_transactions_list"
	"math/big"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrEmptyUserPrivateKey     = errors.New("user private key is empty")
	ErrMoreThenZeroLimit       = errors.New("limit must be more than zero")
	ErrInvalidLimit            = errors.New("limit must be valid value")
	ErrInvalidStartBlockNumber = errors.New("start block number must be valid")
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
) txTransfersList.UseCaseTxTransactionsList {
	return &uc{
		cfg: cfg,
		log: log,
		sb:  sb,
	}
}

func (u *uc) Do(
	ctx context.Context,
	args []string,
	startBlockNumber, limit, userEthPrivateKey string,
) (json.RawMessage, error) {
	const (
		hName               = "UseCase TxTransactionsList"
		senderKey           = "sender"
		limitKey            = "limit"
		startBlockNumberKey = "start_block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(limitKey, limit),
			attribute.String(startBlockNumberKey, startBlockNumber),
		))
	defer span.End()

	if userEthPrivateKey == "" {
		open_telemetry.MarkSpanError(spanCtx, ErrEmptyUserPrivateKey)
		return nil, ErrEmptyUserPrivateKey
	}

	lm, err := strconv.Atoi(limit)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, errors.Join(ErrInvalidLimit, err))
		return nil, ErrInvalidLimit
	}

	if lm < 0 {
		open_telemetry.MarkSpanError(spanCtx, ErrMoreThenZeroLimit)
		return nil, ErrMoreThenZeroLimit
	}

	var (
		bn      *big.Int
		bnCheck bool
	)
	bn, bnCheck = new(big.Int).SetString(startBlockNumber, 10)
	if !bnCheck {
		open_telemetry.MarkSpanError(spanCtx, ErrInvalidStartBlockNumber)
		return nil, ErrInvalidStartBlockNumber
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

	return service.TransactionsList(spanCtx, u.cfg, bn.Uint64(), uint64(lm), userEthPrivateKey)
}
