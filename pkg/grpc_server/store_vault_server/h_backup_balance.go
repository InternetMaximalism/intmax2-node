package store_vault_server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) BackupBalance(ctx context.Context, req *node.BackupBalanceRequest) (*node.BackupBalanceResponse, error) {
	resp := node.BackupBalanceResponse{}

	const (
		hName      = "Handler BackupBalance"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	resp.Success = true
	resp.Data = &node.BackupBalanceResponse_Data{Message: "Backup balance success"}

	return &resp, utils.OK(spanCtx)
}
