package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) BackupDeposit(ctx context.Context, req *node.BackupDepositRequest) (*node.BackupDepositResponse, error) {
	resp := node.BackupDepositResponse{}

	const (
		hName      = "Handler BackupDeposit"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := backupDeposit.UCPostBackupDepositInput{
		EncryptedDeposit: req.EncryptedDeposit,
		Recipient:        req.Recipient,
		BlockNumber:      uint32(req.BlockNumber),
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		err = s.commands.PostBackupDeposit(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to post backup deposit: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	})
	if err != nil {
		const msg = "failed to post backup deposit with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.BackupDepositResponse_Data{Message: backupDeposit.SuccessMsg}

	return &resp, utils.OK(spanCtx)
}
