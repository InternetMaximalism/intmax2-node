package server

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	"intmax2-node/internal/use_cases/backup_balance"
	"intmax2-node/pkg/grpc_server/utils"
)

func (s *Server) BackupBalance(
	ctx context.Context,
	req *node.BackupBalanceRequest,
) (*node.BackupBalanceResponse, error) {
	resp := node.BackupBalanceResponse{}

	const (
		hName      = "Handler Backup Balance"
		requestKey = "request"
	)
	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName, trace.WithAttributes(
		attribute.String(requestKey, req.String()),
	))
	defer span.End()

	input := backup_balance.UCPostBackupBalanceInput{
		User:        req.GetUser(),
		BlockNumber: req.GetBlockNumber(),
		EncryptedBalanceProof: backup_balance.EncryptedPlonky2Proof{
			Proof:                 req.GetEncryptedBalanceProof().GetProof(),
			EncryptedPublicInputs: req.GetEncryptedBalanceProof().GetPublicInputs(),
		},
		EncryptedBalanceData: req.GetEncryptedBalanceData(),
		EncryptedTxs:         req.GetEncryptedTxs(),
		EncryptedTransfers:   req.GetEncryptedTransfers(),
		EncryptedDeposits:    req.GetEncryptedDeposits(),
		Signature:            req.GetSignature(),
	}

	if err := input.Valid(); err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var res *backup_balance.UCPostBackupBalance
	err := s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		res, err = s.commands.BackupBalance(q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to backup user balance: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	})

	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to backup user balance with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Message = res.Message

	return &resp, utils.OK(spanCtx)
}
