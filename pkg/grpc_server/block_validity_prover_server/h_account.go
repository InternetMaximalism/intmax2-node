package block_validity_prover_server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	"intmax2-node/internal/use_cases/block_validity_prover_account"
	"intmax2-node/pkg/grpc_server/utils"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	ucBlockValidityProverAccount "intmax2-node/pkg/use_cases/block_validity_prover_account"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) Account(
	ctx context.Context,
	req *node.AccountRequest,
) (*node.AccountResponse, error) {
	resp := node.AccountResponse{
		Data: &node.AccountResponse_Data{},
	}

	const (
		hName      = "Handler Account"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := block_validity_prover_account.UCBlockValidityProverAccountInput{
		Address: req.Address,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var info block_validity_prover_account.UCBlockValidityProverAccount
	err = s.dbApp.Exec(spanCtx, &info, func(d interface{}, in interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		var result *block_validity_prover_account.UCBlockValidityProverAccount
		result, err = ucBlockValidityProverAccount.New(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			return err
		}

		if v, ok := in.(*block_validity_prover_account.UCBlockValidityProverAccount); ok {
			v.AccountID = result.AccountID
		} else {
			const msg = "failed to convert of account info"
			err = fmt.Errorf(msg)
			open_telemetry.MarkSpanError(spanCtx, err)
			return err
		}

		return nil
	})
	if err == nil {
		resp.Data.IsRegistered = true
		resp.Data.AccountId = uint32(info.AccountID.Uint64())
	} else if !errors.Is(err, errorsDB.ErrNotFound) &&
		!errors.Is(err, ucBlockValidityProverAccount.ErrNewAddressFromHexFail) &&
		!errors.Is(err, ucBlockValidityProverAccount.ErrPublicKeyFromIntMaxAccFail) {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get account by address: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true

	return &resp, utils.OK(spanCtx)
}
