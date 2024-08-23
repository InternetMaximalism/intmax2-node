package server

import (
	"context"
	"intmax2-node/configs/buildvars"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Server) Version(ctx context.Context, _ *node.VersionRequest) (*node.VersionResponse, error) {
	const (
		hName     = "Handler Version"
		version   = "version"
		buildTime = "build_time"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(version, buildvars.Version),
			attribute.String(buildTime, buildvars.BuildTime),
		))
	defer span.End()

	info := s.Commands().GetVersion(buildvars.Version, buildvars.BuildTime).Do(spanCtx)
	return &node.VersionResponse{
		Version:   info.Version,
		Buildtime: info.BuildTime,
	}, utils.OK(spanCtx)
}
