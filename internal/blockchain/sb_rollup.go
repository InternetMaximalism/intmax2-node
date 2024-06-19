package blockchain

import (
	"context"
	"errors"
	"fmt"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/open_telemetry"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (sb *serviceBlockchain) callContractRollup(
	ctx context.Context,
	method string,
	args ...any,
) (resp []interface{}, err error) {
	const (
		hName       = "ServiceBlockchain func:callContractRollup"
		methodKey   = "method"
		argsKey     = "args"
		maskArgsKey = "%+v"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(methodKey, method),
			attribute.StringSlice(argsKey, func() (ret []string) {
				for key := range args {
					ret = append(ret, fmt.Sprintf(maskArgsKey, args[key]))
				}
				return ret
			}()),
		))
	defer span.End()

	for {
		resp, err = sb.callContract(
			spanCtx,
			common.HexToAddress(sb.cfg.Blockchain.RollupContractAddress),
			sb.cfg.Blockchain.TemplateContractRollupPath,
			method,
			args...,
		)
		if err != nil {
			if strings.Contains(err.Error(), errorsB.Err520ScrollWebServerStr) {
				<-time.After(time.Second)
				continue
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return nil, errors.Join(errorsB.ErrCallContractFail, err)
		}
		break
	}

	return resp, nil
}

func (sb *serviceBlockchain) transactorOfContractRollup(
	ctx context.Context,
	value *big.Int,
	method string,
	args ...any,
) (resp *types.Transaction, err error) {
	const (
		hName       = "ServiceBlockchain func:transactOfContractRollup"
		methodKey   = "method"
		argsKey     = "args"
		valueKey    = "value"
		maskArgsKey = "%+v"
		int0StrKey  = "0"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(methodKey, method),
			attribute.StringSlice(argsKey, func() (ret []string) {
				for key := range args {
					ret = append(ret, fmt.Sprintf(maskArgsKey, args[key]))
				}
				return ret
			}()),
		))
	defer span.End()

	if value == nil {
		span.SetAttributes(attribute.String(valueKey, int0StrKey))
	} else {
		span.SetAttributes(attribute.String(valueKey, value.String()))
	}

	for {
		resp, err = sb.contractTransactor(
			spanCtx,
			common.HexToAddress(sb.cfg.Blockchain.RollupContractAddress),
			sb.cfg.Blockchain.TemplateContractRollupPath,
			value,
			method,
			args...,
		)
		if err != nil {
			if strings.Contains(err.Error(), errorsB.Err520ScrollWebServerStr) ||
				strings.Contains(err.Error(), errorsB.ErrInvalidSequenceStr) {
				<-time.After(time.Second)
				continue
			}
			if strings.Contains(err.Error(), errorsB.ErrInsufficientFundsStr) {
				errorsB.InsufficientFunds = true
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return nil, errors.Join(errorsB.ErrApplyContractTransactorFail, err)
		}
		errorsB.InsufficientFunds = false
		break
	}

	return resp, nil
}

func (sb *serviceBlockchain) UpdateBlockBuilder(
	ctx context.Context,
	url string,
) (err error) {
	const (
		hName      = "ServiceBlockchain func:UpdateBlockBuilder"
		urlKey     = "url"
		valueKey   = "value"
		int0StrKey = "0"
		int1Key    = 1
		methodKey  = "updateBlockBuilder"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(urlKey, url),
		))
	defer span.End()

	var (
		value  *big.Int
		tryNum int
	)
	defer func() {
		if value == nil {
			span.SetAttributes(attribute.String(valueKey, int0StrKey))
		} else {
			span.SetAttributes(attribute.String(valueKey, value.String()))
		}
	}()
	for {
		_, err = sb.transactorOfContractRollup(
			spanCtx,
			value,
			methodKey,
			url,
		)
		if err != nil {
			if strings.Contains(err.Error(), errorsB.ErrInsufficientStakeAmountStr) {
				value = &sb.cfg.Blockchain.ScrollNetworkStakeBalance
				if tryNum < int1Key {
					tryNum++
					continue
				}
				errorsB.InsufficientFunds = true
			}

			open_telemetry.MarkSpanError(spanCtx, err)
			return errors.Join(errorsB.ErrApplyTransactOfContractRollupFail, err)
		}
		break
	}

	return nil
}

func (sb *serviceBlockchain) StopBlockBuilder(
	ctx context.Context,
) (err error) {
	const (
		hName     = "ServiceBlockchain func:StopBlockBuilder"
		methodKey = "stopBlockBuilder"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	_, err = sb.transactorOfContractRollup(
		spanCtx,
		nil,
		methodKey,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrApplyTransactOfContractRollupFail, err)
	}

	return nil
}
