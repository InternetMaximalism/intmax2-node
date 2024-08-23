package server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/internal/use_cases/block_status"
	"intmax2-node/pkg/grpc_server/utils"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	base10        = 10
	numUint64Bits = 64
)

func (s *Server) BlockStatusByTxTreeRoot(
	ctx context.Context,
	req *node.BlockStatusByTxTreeRootRequest,
) (*node.BlockStatusByTxTreeRootResponse, error) {
	resp := node.BlockStatusByTxTreeRootResponse{}

	const (
		hName      = "Handler BlockStatusByTxTreeRoot"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := block_status.UCBlockStatusInput{
		TxTreeRoot: req.TxTreeRoot,
	}

	blockStatus, err := s.commands.BlockStatusByTxTreeRoot(s.config, s.log, s.dbApp, s.worker).Do(spanCtx, &input)
	if err != nil {
		if err.Error() == "not found" || err.Error() == "tx tree root not found" || err.Error() == "transaction hash not found" {
			return &resp, utils.NotFound(spanCtx, err)
		}

		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get block status: %v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.IsPosted = blockStatus.IsPosted
	resp.BlockNumber, err = strconv.ParseUint(blockStatus.BlockNumber, base10, numUint64Bits)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
	}

	return &resp, utils.OK(spanCtx)
}
