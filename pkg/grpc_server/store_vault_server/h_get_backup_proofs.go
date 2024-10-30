package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	backupProof "intmax2-node/internal/use_cases/backup_balance_proof"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) GetBackupBalanceProofs(ctx context.Context, req *node.GetBackupBalanceProofsRequest) (*node.GetBackupBalanceProofsResponse, error) {
	resp := node.GetBackupBalanceProofsResponse{}

	const (
		hName      = "Handler GetBackupBalances"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := backupProof.UCGetBackupBalanceProofsInput{
		Hashes: req.Hashes,
	}

	// err := input.Valid()
	// if err != nil {
	// 	open_telemetry.MarkSpanError(spanCtx, err)
	// 	return &resp, utils.BadRequest(spanCtx, err)
	// }

	err := s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		results, err := s.commands.GetBackupSenderBalanceProofs(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get backup balances: %w"
			return fmt.Errorf(msg, err)
		}
		resp.Data = results

		return nil
	})
	if err != nil {
		const msg = "failed to get backup balances with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true

	return &resp, utils.OK(spanCtx)
}
