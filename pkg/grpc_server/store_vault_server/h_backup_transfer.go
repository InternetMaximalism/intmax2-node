package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) BackupTransfer(ctx context.Context, req *node.BackupTransferRequest) (*node.BackupTransferResponse, error) {
	resp := node.BackupTransferResponse{}

	const (
		hName      = "Handler BackupTransfer"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := backupTransfer.UCPostBackupTransferInput{
		EncryptedTransfer: req.EncryptedTransfer,
		Recipient:         req.Recipient,
		BlockNumber:       uint32(req.BlockNumber),
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		err = s.commands.PostBackupTransfer(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to post backup transfer: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	})
	if err != nil {
		const msg = "failed to post backup transfer with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.BackupTransferResponse_Data{Message: backupTransfer.SuccessMsg}

	return &resp, utils.OK(spanCtx)
}
