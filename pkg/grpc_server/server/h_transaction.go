package server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/internal/worker"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Server) Transaction(
	ctx context.Context,
	req *node.TransactionRequest,
) (*node.TransactionResponse, error) {
	resp := node.TransactionResponse{}

	const (
		hName      = "Handler Transaction"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName, trace.WithAttributes(
		attribute.String(requestKey, req.String()),
	))
	defer span.End()

	input := transaction.UCTransactionInput{
		Sender:        req.Sender,
		TransfersHash: req.TransfersHash,
		Nonce:         req.Nonce,
		PowNonce:      req.PowNonce,
		Expiration:    req.Expiration.AsTime(),
		Signature:     req.Signature,
	}

	for key := range req.TransferData {
		data := transaction.TransferDataTransaction{
			TokenIndex: req.TransferData[key].TokenIndex,
			Amount:     req.TransferData[key].Amount,
			Salt:       req.TransferData[key].Salt,
		}
		if req.TransferData[key].Recipient != nil {
			data.Recipient = &transaction.RecipientTransferDataTransaction{
				AddressType: req.TransferData[key].Recipient.AddressType.String(),
				Address:     req.TransferData[key].Recipient.Address,
			}
		}
		input.TransferData = append(input.TransferData, &data)
	}

	err := input.Valid(s.config, s.pow)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.commands.Transaction(s.config, s.worker).Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		if errors.Is(err, worker.ErrReceiverWorkerDuplicate) {
			const msg = "%s"
			return &resp, utils.BadRequest(spanCtx, fmt.Errorf(msg, transaction.NotUniqueMsg))
		}

		const msg = "failed to commit transaction: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.DataTransactionResponse{Message: transaction.SuccessMsg}

	return &resp, utils.OK(spanCtx)
}
