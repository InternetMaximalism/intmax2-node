package blockchain

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/open_telemetry"
	"os"
	"strings"

	"github.com/prodadidb/go-validation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var ErrScrollChainIDInvalid = fmt.Errorf(errorsB.ErrScrollChainIDInvalidStr, ScrollMainNetChainID, ScrollSepoliaChainID)

type ChainIDType string

const (
	ScrollMainNetChainID ChainIDType = "534352"
	ScrollSepoliaChainID ChainIDType = "534351"
)

type ChainLinkEvmJSONRPC string

const (
	ScrollMainNetChainLinkEvmJSONRPC ChainLinkEvmJSONRPC = "https://rpc.scroll.io"
	ScrollSepoliaChainLinkEvmJSONRPC ChainLinkEvmJSONRPC = "https://sepolia-rpc.scroll.io"
)

type ChainLinkExplorer string

const (
	ScrollMainNetChainLinkExplorer ChainLinkExplorer = "https://sepolia.scrollscan.com"
	ScrollSepoliaChainLinkExplorer ChainLinkExplorer = "https://scrollscan.com"
)

func (sb *serviceBlockchain) scrollNetworkChainIDValidator() error {
	return validation.Validate(sb.cfg.Blockchain.ScrollNetworkChainID,
		validation.Required,
		validation.In(
			string(ScrollMainNetChainID), string(ScrollSepoliaChainID),
		),
	)
}

func (sb *serviceBlockchain) SetupScrollNetworkChainID(ctx context.Context) error {
	const (
		hName = "ServiceBlockchain func:SetupScrollNetworkChainID"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	err := sb.scrollNetworkChainIDValidator()
	if err != nil {
		const (
			enterMSG = "Enter the Scroll network chain-ID:"
			crlf     = '\n'
		)
		fmt.Printf(enterMSG)
		var chainID string
		chainID, err = bufio.NewReader(os.Stdin).ReadString(crlf)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return errors.Join(errorsB.ErrStdinProcessingFail, err)
		}
		sb.cfg.Blockchain.ScrollNetworkChainID = strings.TrimSpace(chainID)
	}

	err = sb.scrollNetworkChainIDValidator()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrScrollChainIDInvalid, err)
	}

	return nil
}

func (sb *serviceBlockchain) ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error) {
	const (
		hName                   = "ServiceBlockchain func:ScrollNetworkChainLinkEvmJSONRPC"
		scrollNetworkChainIDKey = "scroll_network_chain_id"
		emptyKey                = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(scrollNetworkChainIDKey, sb.cfg.Blockchain.ScrollNetworkChainID),
		))
	defer span.End()

	err := sb.scrollNetworkChainIDValidator()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return emptyKey, errors.Join(ErrScrollChainIDInvalid, err)
	}

	if strings.EqualFold(sb.cfg.Blockchain.ScrollNetworkChainID, string(ScrollMainNetChainID)) {
		return string(ScrollMainNetChainLinkEvmJSONRPC), nil
	}

	return string(ScrollSepoliaChainLinkEvmJSONRPC), nil
}

func (sb *serviceBlockchain) ScrollNetworkChainLinkExplorer(ctx context.Context) (string, error) {
	const (
		hName                   = "ServiceBlockchain func:ScrollNetworkChainLinkExplorer"
		scrollNetworkChainIDKey = "scroll_network_chain_id"
		emptyKey                = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(scrollNetworkChainIDKey, sb.cfg.Blockchain.ScrollNetworkChainID),
		))
	defer span.End()

	err := sb.scrollNetworkChainIDValidator()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return emptyKey, errors.Join(ErrScrollChainIDInvalid, err)
	}

	if strings.EqualFold(sb.cfg.Blockchain.ScrollNetworkChainID, string(ScrollMainNetChainID)) {
		return string(ScrollMainNetChainLinkExplorer), nil
	}

	return string(ScrollSepoliaChainLinkExplorer), nil
}
