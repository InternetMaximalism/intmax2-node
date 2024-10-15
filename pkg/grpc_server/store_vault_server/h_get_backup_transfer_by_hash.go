package store_vault_server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	getBackupTransferByHash "intmax2-node/internal/use_cases/get_backup_transfer_by_hash"
	"intmax2-node/pkg/grpc_server/utils"
	errorsDB "intmax2-node/pkg/sql_db/errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) GetBackupTransferByHash(
	ctx context.Context,
	req *node.GetBackupTransferByHashRequest,
) (*node.GetBackupTransferByHashResponse, error) {
	resp := node.GetBackupTransferByHashResponse{}

	const (
		hName      = "Handler GetBackupTransferByHash"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := getBackupTransferByHash.UCGetBackupTransferByHashInput{
		Recipient:    req.Recipient,
		TransferHash: req.TransferHash,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		results, err := s.commands.GetBackupTransferByHash(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to get backup transfer: %w"
			return fmt.Errorf(msg, err)
		}
		resp.Data = results

		return nil
	})
	if err != nil {
		if errors.Is(err, errorsDB.ErrNotFound) {
			return &resp, utils.NotFound(spanCtx, fmt.Errorf("%s", getBackupTransferByHash.NotFoundMessage))
		}

		const msg = "failed to get backup transfer with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true

	return &resp, utils.OK(spanCtx)
}
