package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) GetBackupBalances(ctx context.Context, req *node.GetBackupBalancesRequest) (*node.GetBackupBalancesResponse, error) {
	fmt.Printf("GetBackupBalances: %v", req)
	panic("GetBackupBalances")
	resp := node.GetBackupBalancesResponse{}

	const (
		hName      = "Handler GetBackupBalances"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := backupBalance.UCGetBackupBalancesInput{
		Sender:           req.Sender,
		StartBlockNumber: req.StartBlockNumber,
		Limit:            req.Limit,
	}

	err := input.Valid()
	if err != nil {
		fmt.Printf("Error: %v", err)
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		results, err := s.commands.GetBackupBalances(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			fmt.Printf("Error: %v", err)
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get backup balances: %w"
			return fmt.Errorf(msg, err)
		}
		resp.Data = results

		return nil
	})
	if err != nil {
		fmt.Printf("Error: %v", err)
		const msg = "failed to get backup balances with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true

	return &resp, utils.OK(spanCtx)
}
