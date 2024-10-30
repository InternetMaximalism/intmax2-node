package block_validity_prover_server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) BlockValidityProverInfo(
	ctx context.Context,
	req *node.BlockValidityProverInfoRequest,
) (*node.BlockValidityProverInfoResponse, error) {
	resp := node.BlockValidityProverInfoResponse{}

	const (
		hName      = "Handler BlockValidityProverInfo"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	info, err := s.Commands().BlockValidityProverInfo(s.config, s.log, s.bvs).Do(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get info of block validity prover: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.BlockValidityProverInfoResponse_Data{
		DepositIndex: info.DepositIndex,
		BlockNumber:  info.BlockNumber,
	}

	return &resp, utils.OK(spanCtx)
}
