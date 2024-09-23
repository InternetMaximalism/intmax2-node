package block_validity_prover_server

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	"intmax2-node/pkg/grpc_server/utils"
)

func (s *BlockValidityProverServer) DepositTreeProof(
	ctx context.Context,
	req *node.DepositTreeProofRequest,
) (*node.DepositTreeProofResponse, error) {
	const (
		hName      = "Handler DepositTreeProof"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	resp := node.DepositTreeProofResponse{}

	return &resp, utils.OK(spanCtx)
}
