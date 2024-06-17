package blockchain

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/mnemonic_wallet"
	modelsMW "intmax2-node/internal/mnemonic_wallet/models"
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

func (sb *serviceBlockchain) BlockBuilderUrl(ctx context.Context) (url string, err error) {
	const (
		hName      = "ServiceBlockchain func:BlockBuilderUrl"
		addressKey = "address"
		methodKey  = "blockBuilderUrl"
		emptyKey   = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	var pk string
	pk, err = sb.recognizingPrivateKey(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return emptyKey, errors.Join(errorsB.ErrRecognizingPrivateKeyFail, err)
	}

	var w *modelsMW.Wallet
	w, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(pk)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return emptyKey, errors.Join(errorsB.ErrWalletAddressNotRecognized, err)
	}

	span.SetAttributes(attribute.String(addressKey, w.WalletAddress.String()))

	var data []interface{}
	data, err = sb.callContractRollup(spanCtx, methodKey, w.WalletAddress)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return emptyKey, errors.Join(errorsB.ErrCallRollupContractFail, err)
	}

	resp := *abi.ConvertType(data[0], new(string)).(*string)

	return strings.TrimSpace(resp), nil
}

func (sb *serviceBlockchain) UpdateBlockBuilder(
	ctx context.Context,
	url string,
) (err error) {
	const (
		hName     = "ServiceBlockchain func:UpdateBlockBuilder"
		urlKey    = "url"
		methodKey = "updateBlockBuilder"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(urlKey, url),
		))
	defer span.End()

	_, err = sb.transactorOfContractRollup(
		spanCtx,
		nil,
		methodKey,
		url,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrApplyTransactOfContractRollupFail, err)
	}

	return nil
}
