package server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/internal/use_cases/deposit_status_by_hash"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Server) DepositStatusByHash(
	ctx context.Context,
	req *node.DepositStatusByHashRequest,
) (resp *node.DepositStatusByHashResponse, err error) {
	const (
		hName      = "Handler DepositStatusByHash"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := deposit_status_by_hash.UCDepositStatusByHashInput{
		DepositHash: req.DepositHash,
	}

	// err := input.Valid(s.worker)
	// if err != nil {
	// 	open_telemetry.MarkSpanError(spanCtx, err)
	// 	return &resp, utils.BadRequest(spanCtx, err)
	// }

	var ucBP *deposit_status_by_hash.UCDepositStatusByHash
	ucBP, err = s.commands.DepositStatusByHash(
		s.config,
		s.log,
		s.dbApp,
		s.blockValidityProver,
	).Do(spanCtx, &input)
	if err != nil {
		// open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get block proposed: %v"
		return resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.DataDepositStatusByHashResponse{
		BlockNumber: ucBP.BlockNumber,
	}

	return resp, utils.OK(spanCtx)
}
