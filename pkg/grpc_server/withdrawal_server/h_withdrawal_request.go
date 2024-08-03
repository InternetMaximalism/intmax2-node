package withdrawal_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/blockchain"
	"intmax2-node/internal/open_telemetry"
	"intmax2-node/internal/pb/gen/service/node"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	"intmax2-node/pkg/grpc_server/utils"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *WithdrawalServer) WithdrawalRequest(ctx context.Context, req *node.WithdrawalRequestRequest) (*node.WithdrawalRequestResponse, error) {
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
		TransferData: mDBApp.TransferData{
			Recipient:  req.TransferData.Recipient,
			TokenIndex: req.TransferData.TokenIndex,
			Amount:     req.TransferData.Amount,
			Salt:       req.TransferData.Salt,
		},
		TransferMerkleProof: mDBApp.TransferMerkleProof{
			Siblings: req.TransferMerkleProof.Siblings,
			Index:    req.TransferMerkleProof.Index,
		},
		Transaction: mDBApp.Transaction{
			TransferTreeRoot: req.Transaction.TransferTreeRoot,
			Nonce:            req.Transaction.Nonce,
		},
		TxMerkleProof: mDBApp.TxMerkleProof{
			Siblings: req.TxMerkleProof.Siblings,
			Index:    req.TxMerkleProof.Index,
		},
		TransferHash: req.TransferHash,
		BlockNumber:  req.BlockNumber,
		BlockHash:    req.BlockHash,
		EnoughBalanceProof: mDBApp.EnoughBalanceProof{
			Proof:        req.EnoughBalanceProof.Proof,
			PublicInputs: req.EnoughBalanceProof.PublicInputs,
		},
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	bc := blockchain.New(ctx, s.config)

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
		const msg = "failed to post withdrawal request with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.WithdrawalRequestResponse_Data{Message: "Withdraw request accepted."}

	return &resp, utils.OK(spanCtx)
}
