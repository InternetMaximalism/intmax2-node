package server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/pkg/grpc_server/utils"
)

func (s *Server) Info(
	ctx context.Context,
	_ *node.InfoRequest,
) (*node.InfoResponse, error) {
	resp := node.InfoResponse{}

	const (
		hName = "Handler Info"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	info, err := s.commands.BlockInfo(s.config, s.storageGPO).Do(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get block info: %v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Data = &node.DataInfoResponse{
		TransferFee:   info.TransferFee,
		Difficulty:    uint32(info.Difficulty),
		ScrollAddress: info.ScrollAddress,
		IntMaxAddress: info.IntMaxAddress,
	}

	resp.Success = true

	return &resp, utils.OK(spanCtx)
}
