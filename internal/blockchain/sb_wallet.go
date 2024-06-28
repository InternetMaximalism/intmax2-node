package blockchain

import (
	"context"
	"errors"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/open_telemetry"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (sb *serviceBlockchain) walletBalance(
	ctx context.Context,
	address common.Address,
) (bal *big.Int, err error) {
	const (
		hName            = "ServiceBlockchain func:walletBalance"
		walletAddressKey = "wallet_address"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(walletAddressKey, address.String()),
		))
	defer span.End()

	var (
		c      *ethclient.Client
		cancel func()
	)
	c, cancel, err = sb.ethClient(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(errorsB.ErrCreateEthClientFail, err)
	}
	defer cancel()

	var bn uint64
	for {
		bn, err = c.BlockNumber(spanCtx)
		if err != nil {
			if strings.Contains(err.Error(), errorsB.Err520ScrollWebServerStr) ||
				strings.Contains(err.Error(), errorsB.Err502ScrollWebServerStr) {
				<-time.After(time.Second)
				continue
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return nil, errors.Join(errorsB.ErrMostBlockNumberFail, err)
		}
		break
	}

	for {
		bal, err = c.BalanceAt(spanCtx, address, new(big.Int).SetUint64(bn))
		if err != nil {
			if strings.Contains(err.Error(), errorsB.Err520ScrollWebServerStr) ||
				strings.Contains(err.Error(), errorsB.Err502ScrollWebServerStr) {
				<-time.After(time.Second)
				continue
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return nil, errors.Join(errorsB.ErrGetWalletBalanceFail, err)
		}
		break
	}

	return bal, nil
}
