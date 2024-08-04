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
	req *node.BlockStatusRequest,
) (*node.BlockStatusResponse, error) {
	resp := node.BlockStatusResponse{}

	const (
		hName      = "Handler BlockSignature"
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

	blockStatus, err := s.commands.BlockStatus(s.config, s.log).Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get block signature: %v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.IsPosted = blockStatus.IsPosted
	resp.BlockNumber = uint64(blockStatus.BlockNumber)

	return &resp, utils.OK(spanCtx)
}
