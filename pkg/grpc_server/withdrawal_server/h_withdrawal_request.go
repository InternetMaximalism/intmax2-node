package withdrawal_server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/internal/blockchain"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/withdrawal_service/node"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	withdrawalService "intmax2-node/internal/withdrawal_service"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *WithdrawalServer) WithdrawalRequest(
	ctx context.Context,
	req *node.WithdrawalRequestRequest,
) (*node.WithdrawalRequestResponse, error) {
	resp := node.WithdrawalRequestResponse{}

	const (
		hName      = "Handler WithdrawalRequest"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := postWithdrwalRequest.UCPostWithdrawalRequestInput{
		TransferData: &postWithdrwalRequest.UCPostWithdrawalRequestTransferDataInput{
			Recipient:  req.TransferData.Recipient,
			TokenIndex: int64(req.TransferData.TokenIndex),
			Amount:     req.TransferData.Amount,
			Salt:       req.TransferData.Salt,
		},
		TransferMerkleProof: &postWithdrwalRequest.UCPostWithdrawalRequestTransferMerkleProofInput{
			Siblings: req.TransferMerkleProof.Siblings,
			Index:    int64(req.TransferMerkleProof.Index),
		},
		Transaction: &postWithdrwalRequest.UCPostWithdrawalRequestTransactionInput{
			TransferTreeRoot: req.Transaction.TransferTreeRoot,
			Nonce:            int64(req.Transaction.Nonce),
		},
		TxMerkleProof: &postWithdrwalRequest.UCPostWithdrawalRequestTxMerkleProofInput{
			Siblings: req.TxMerkleProof.Siblings,
			Index:    int64(req.TxMerkleProof.Index),
		},
		TransferHash: req.TransferHash,
		BlockNumber:  int64(req.BlockNumber),
		BlockHash:    req.BlockHash,
		EnoughBalanceProof: &postWithdrwalRequest.UCPostWithdrawalRequestEnoughBalanceProofInput{
			Proof:        req.EnoughBalanceProof.Proof,
			PublicInputs: req.EnoughBalanceProof.PublicInputs,
		},
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	bc := blockchain.New(ctx, s.config, s.log)

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		err = s.commands.PostWithdrawalRequest(s.config, s.log, q, bc).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to post withdrawal request: %w"
			return fmt.Errorf(msg, err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, withdrawalService.ErrWithdrawalRequestAlreadyExists) {
			return &resp, utils.BadRequest(spanCtx, withdrawalService.ErrWithdrawalRequestAlreadyExists)
		}

		const msg = "failed to post withdrawal request with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.WithdrawalRequestResponse_Data{Message: postWithdrwalRequest.SuccessMsg}

	return &resp, utils.OK(spanCtx)
}
