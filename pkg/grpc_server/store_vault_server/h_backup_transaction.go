package store_vault_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	postBackupTransction "intmax2-node/internal/use_cases/post_backup_transaction"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *StoreVaultServer) BackupTransaction(
	ctx context.Context,
	req *node.BackupTransactionRequest,
) (*node.BackupTransactionResponse, error) {
	resp := node.BackupTransactionResponse{}

	const (
		hName      = "Handler BackupTransaction"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := postBackupTransction.UCPostBackupTransactionInput{
		TxHash:      req.TxHash,
		EncryptedTx: req.EncryptedTx,
		Sender:      req.Sender,
		Signature:   req.Signature,
		BlockNumber: uint32(req.BlockNumber),
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		err = s.commands.PostBackupTransaction(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to post backup transaction: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	})
	if err != nil {
		const msg = "failed to post backup transaction with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.BackupTransactionResponse_Data{Message: postBackupTransction.SuccessMsg}

	return &resp, utils.OK(spanCtx)
}
