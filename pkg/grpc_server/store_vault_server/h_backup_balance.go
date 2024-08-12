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

	input := backupBalance.UCPostBackupBalanceInput{
		User:                  req.User,
		EncryptedBalanceProof: req.EncryptedBalanceProof,
		EncryptedBalanceData:  req.EncryptedBalanceData,
		EncryptedTxs:          req.EncryptedTxs,
		EncryptedTransfers:    req.EncryptedTransfers,
		EncryptedDeposits:     req.EncryptedDeposits,
		Signature:             req.Signature,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		err = s.commands.PostBackupBalance(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to post backup balance: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	})
	if err != nil {
		const msg = "failed to post backup balance with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.BackupBalanceResponse_Data{Message: backupBalance.SuccessMsg}

	return &resp, utils.OK(spanCtx)
}
