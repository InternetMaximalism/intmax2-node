package server

// import (
// 	"context"
// 	"intmax2-node/internal/open_telemetry"
// 	node "intmax2-node/internal/pb/gen/block_builder_service/node"
// 	"intmax2-node/internal/use_cases/block_number_by_tx_root"
// 	"intmax2-node/pkg/grpc_server/utils"

// 	"go.opentelemetry.io/otel/attribute"
// 	"go.opentelemetry.io/otel/trace"
// )

// func (s *Server) BlockNumberByTxHash(
// 	ctx context.Context,
// 	req *node.BlockNumberByTxHashRequest,
// ) (resp *node.BlockNumberByTxHashResponse, err error) {
// 	const (
// 		hName      = "Handler BlockNumberByTxRoot"
// 		requestKey = "request"
// 	)

// 	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
// 		trace.WithAttributes(
// 			attribute.String(requestKey, req.String()),
// 		))
// 	defer span.End()

// 	input := block_number_by_tx_root.UCBlockNumberByTxRootInput{
// 		TxRoot: req.TxRoot,
// 	}

// 	// err := input.Valid(s.worker)
// 	// if err != nil {
// 	// 	open_telemetry.MarkSpanError(spanCtx, err)
// 	// 	return &resp, utils.BadRequest(spanCtx, err)
// 	// }

// 	var ucBP *block_number_by_tx_hash.UCBlockNumberByTxHash
// 	ucBP, err = s.commands.BlockNumberByTxHash().Do(spanCtx, &input)
// 	if err != nil {
// 		// open_telemetry.MarkSpanError(spanCtx, err)
// 		const msg = "failed to get block proposed: %v"
// 		return &resp, utils.Internal(spanCtx, s.log, msg, err)
// 	}

// 	resp.Success = true
// 	resp.Data = &node.DataBlockNumberByTxHashResponse{
// 		TxRoot: ucBP.BlockNumber,
// 	}

// 	return &resp, utils.OK(spanCtx)
// }
