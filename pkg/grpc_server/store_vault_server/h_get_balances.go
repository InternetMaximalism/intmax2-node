package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) GetBalances(ctx context.Context, req *node.GetBalancesRequest) (*node.GetBalancesResponse, error) {
	resp := node.GetBalancesResponse{}

	const (
		hName      = "Handler GetBalances"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := backupBalance.UCGetBalancesInput{
		Address: req.Address,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		results, err := s.commands.GetBalances(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get balances request: %w"
			return fmt.Errorf(msg, err)
		}
		resp.Balances = convertToTokenBalances(results.Balances)
		return nil
	})
	if err != nil {
		const msg = "failed to get balances with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	return &resp, utils.OK(spanCtx)
}

func convertToTokenBalances(balances []*backupBalance.TokenBalance) []*node.TokenBalance {
	result := make([]*node.TokenBalance, len(balances))
	for i, balance := range balances {
		result[i] = &node.TokenBalance{
			TokenIndex: int32(balance.TokenIndex),
			Amount:     balance.Amount,
		}
	}
	return result
}
