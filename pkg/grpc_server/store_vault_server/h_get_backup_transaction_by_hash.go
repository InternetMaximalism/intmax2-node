package store_vault_server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	getBackupTransactionByHash "intmax2-node/internal/use_cases/get_backup_transaction_by_hash"
	"intmax2-node/pkg/grpc_server/utils"
	errorsDB "intmax2-node/pkg/sql_db/errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) GetBackupTransactionByHash(
	ctx context.Context,
	req *node.GetBackupTransactionByHashRequest,
) (*node.GetBackupTransactionByHashResponse, error) {
	resp := node.GetBackupTransactionByHashResponse{}

	const (
		hName      = "Handler GetBackupTransactions"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := getBackupTransactionByHash.UCGetBackupTransactionByHashInput{
		Sender: req.Sender,
		TxHash: req.TxHash,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		results, err := s.commands.GetBackupTransactionByHash(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get backup transactions: %w"
			return fmt.Errorf(msg, err)
		}
		resp.Data = results

		return nil
	})
	if err != nil {
		if errors.Is(err, errorsDB.ErrNotFound) {
			return &resp, utils.NotFound(spanCtx, fmt.Errorf("%s", getBackupTransactionByHash.NotFoundMessage))
		}

		const msg = "failed to get backup transactions with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true

	return &resp, utils.OK(spanCtx)
}
