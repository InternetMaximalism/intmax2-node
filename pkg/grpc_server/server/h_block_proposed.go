package server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/internal/use_cases/block_proposed"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Server) BlockProposed(
	ctx context.Context,
	req *node.BlockProposedRequest,
) (*node.BlockProposedResponse, error) {
	resp := node.BlockProposedResponse{}

	const (
		hName      = "Handler BlockProposed"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := block_proposed.UCBlockProposedInput{
		Sender:     req.Sender,
		TxHash:     req.TxHash,
		Expiration: req.Expiration.AsTime(),
		Signature:  req.Signature,
	}

	err := input.Valid(s.worker)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var ucBP *block_proposed.UCBlockProposed
	ucBP, err = s.commands.BlockProposed().Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get block proposed: %v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.DataBlockProposedResponse{
		TxRoot:            ucBP.TxRoot,
		TxTreeMerkleProof: make([]string, len(ucBP.TxTreeMerkleProof)),
		PublicKeys:        make([]string, len(ucBP.PublicKeys)),
	}
	copy(resp.Data.TxTreeMerkleProof, ucBP.TxTreeMerkleProof)
	copy(resp.Data.PublicKeys, ucBP.PublicKeys)

	return &resp, utils.OK(spanCtx)
}
