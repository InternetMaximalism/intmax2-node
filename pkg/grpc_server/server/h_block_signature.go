package server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	"intmax2-node/internal/use_cases/block_signature"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Server) BlockSignature(
	ctx context.Context,
	req *node.BlockSignatureRequest,
) (*node.BlockSignatureResponse, error) {
	resp := node.BlockSignatureResponse{}

	const (
		hName      = "Handler BlockSignature"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := block_signature.UCBlockSignatureInput{
		Sender:    req.Sender,
		TxHash:    req.TxHash,
		Signature: req.Signature,
		EnoughBalanceProof: &block_signature.EnoughBalanceProofInput{
			PrevBalanceProof:  &block_signature.Plonky2Proof{},
			TransferStepProof: &block_signature.Plonky2Proof{},
		},
		BackupTx:        &transaction.BackupTransactionData{},
		BackupTransfers: make([]*transaction.BackupTransferInput, len(req.BackupTransfers)),
	}

	if req.BackupTransaction != nil {
		input.BackupTx.EncodedEncryptedTx = req.BackupTransaction.EncryptedTx
		input.BackupTx.Signature = req.BackupTransaction.Signature
	}

	for key := range req.BackupTransfers {
		data := transaction.BackupTransferInput{
			Recipient:                req.BackupTransfers[key].Recipient,
			EncodedEncryptedTransfer: req.BackupTransfers[key].EncryptedTransfer,
		}
		input.BackupTransfers[key] = &data
	}

	if req.EnoughBalanceProof != nil {
		if req.EnoughBalanceProof.PrevBalanceProof != nil {
			input.EnoughBalanceProof.PrevBalanceProof.
				PublicInputs = req.EnoughBalanceProof.PrevBalanceProof.PublicInputs
			input.EnoughBalanceProof.PrevBalanceProof.
				Proof = req.EnoughBalanceProof.PrevBalanceProof.Proof
		}
		if req.EnoughBalanceProof.TransferStepProof != nil {
			input.EnoughBalanceProof.TransferStepProof.
				PublicInputs = req.EnoughBalanceProof.TransferStepProof.PublicInputs
			input.EnoughBalanceProof.TransferStepProof.
				Proof = req.EnoughBalanceProof.TransferStepProof.Proof
		}
	}

	err := input.Valid(s.worker)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.commands.BlockSignature(s.config, s.log, s.worker).Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get block signature: %v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.DataBlockSignatureResponse{
		Message: block_signature.SuccessMsg,
	}

	return &resp, utils.OK(spanCtx)
}
