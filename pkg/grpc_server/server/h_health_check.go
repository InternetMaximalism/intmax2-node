package server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/pkg/grpc_server/utils"
)

func (s *Server) HealthCheck(
	ctx context.Context,
	_ *node.HealthCheckRequest,
) (*node.HealthCheckResponse, error) {
	const (
		hName = "Handler HealthCheck"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	check := s.Commands().HealthCheck(s.hc).Do(spanCtx)
	resp := node.HealthCheckResponse{
		Success: check.Success,
	}

	return &resp, utils.OK(spanCtx)
}
