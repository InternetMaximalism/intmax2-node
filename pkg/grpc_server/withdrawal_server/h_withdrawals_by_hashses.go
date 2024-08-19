package withdrawal_server

import (
	"context"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/withdrawal_service/node"
	postWithdrwalsByHashes "intmax2-node/internal/use_cases/post_withdrawals_by_hashes"
	"intmax2-node/pkg/grpc_server/utils"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *WithdrawalServer) WithdrawalsByHashes(ctx context.Context, req *node.WithdrawalsByHashesRequest) (*node.WithdrawalsByHashesResponse, error) {
	resp := node.WithdrawalsByHashesResponse{}

	const (
		hName      = "Handler WithdrawalsByHashes"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := postWithdrwalsByHashes.UCPostWithdrawalsByHashesInput{
		TransferHashes: req.TransferHashes,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		withdrawals, err := s.commands.PostWithdrawalsByHashes(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			const msg = "failed to post withdrawals by hashes: %w"
			return fmt.Errorf(msg, err)
		}
		for i := range *withdrawals {
			w := (*withdrawals)[i]
			resp.Withdrawals = append(resp.Withdrawals, &node.Withdrawal{
				TransferData: &node.TransferData{
					Recipient:  w.TransferData.Recipient,
					TokenIndex: w.TransferData.TokenIndex,
					Amount:     w.TransferData.Amount,
					Salt:       w.TransferData.Salt,
				},
				Transaction: &node.Transaction{
					TransferTreeRoot: w.Transaction.TransferTreeRoot,
					Nonce:            w.Transaction.Nonce,
				},
				TransferHash: w.TransferHash,
				BlockNumber:  w.BlockNumber,
				BlockHash:    w.BlockHash,
				Status:       mDBApp.WithdrawalStatus(w.Status).String(),
			})
		}

		return nil
	})
	if err != nil {
		const msg = "failed to post withdrawals by hashes with DB App: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	return &resp, utils.OK(spanCtx)
}
