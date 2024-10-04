package store_vault_server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	block_signature "intmax2-node/internal/use_cases/block_signature"
	postBackupTransaction "intmax2-node/internal/use_cases/post_backup_transaction"
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

	if req.SenderEnoughBalanceProofBody == nil {
		const msg = "sender enough balance proof body is nil"
		s.log.Errorf(msg)
		err := errors.New(msg)
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	senderEnoughBalanceProofBody := block_signature.EnoughBalanceProofBodyInput{
		PrevBalanceProofBody:  req.SenderEnoughBalanceProofBody.PrevBalanceProof,
		TransferStepProofBody: req.SenderEnoughBalanceProofBody.TransitionStepProof,
	}
	// senderEnoughBalanceProofBody, err := senderLastBalanceProofBodyInput.EnoughBalanceProofBody()
	// if err != nil {
	// 	s.log.Errorf("failed to get enough balance proof body: %+v\n", err)
	// 	open_telemetry.MarkSpanError(spanCtx, err)
	// 	return &resp, utils.BadRequest(spanCtx, err)
	// }
	input := postBackupTransaction.UCPostBackupTransactionInput{
		TxHash:                       req.TxHash,
		EncryptedTx:                  req.EncryptedTx,
		SenderEnoughBalanceProofBody: &senderEnoughBalanceProofBody,
		Sender:                       req.Sender,
		Signature:                    req.Signature,
		BlockNumber:                  uint32(req.BlockNumber),
		// SenderLastBalanceProofBody:       senderLastBalanceProofBody.PrevBalanceProofBody,
		// SenderBalanceTransitionProofBody: senderLastBalanceProofBody.TransferStepProofBody,
	}

	err := input.Valid()
	if err != nil {
		s.log.Errorf("failed to validate input: %+v\n", err)
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		err = s.commands.PostBackupTransaction(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			s.log.Errorf("failed to post backup transaction: %+v\n", err)
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to post backup transaction: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	})
	if err != nil {
		s.log.Errorf("failed to post backup transaction with DB App: %+v\n", err)
		const msg = "failed to post backup transaction with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.BackupTransactionResponse_Data{Message: postBackupTransaction.SuccessMsg}

	return &resp, utils.OK(spanCtx)
}
