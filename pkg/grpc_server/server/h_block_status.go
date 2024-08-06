package server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	"intmax2-node/internal/use_cases/block_status"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
		if err.Error() == "not found" || err.Error() == "tx tree root not found" {
			return &resp, utils.NotFound(spanCtx, err)
		}

		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get block status: %v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.IsPosted = blockStatus.IsPosted
	resp.BlockNumber = uint64(blockStatus.BlockNumber)

	return &resp, utils.OK(spanCtx)
}
