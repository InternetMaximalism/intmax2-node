package block_validity_prover_server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	"intmax2-node/internal/use_cases/block_validity_prover_block_status_by_block_hash"
	"intmax2-node/pkg/grpc_server/utils"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	ucBlockValidityProverBlockStatusByBlockHash "intmax2-node/pkg/use_cases/block_validity_prover_block_status_by_block_hash"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) BlockHashStatus(
	ctx context.Context,
	req *node.BlockHashStatusRequest,
) (*node.BlockHashStatusResponse, error) {
	resp := node.BlockHashStatusResponse{}

	const (
		hName      = "Handler BlockHashStatus"
		requestKey = "request"
		emptyKey   = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := block_validity_prover_block_status_by_block_hash.UCBlockValidityProverBlockStatusByBlockHashInput{
		BlockHash: strings.TrimSpace(req.BlockHash),
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var info *block_validity_prover_block_status_by_block_hash.UCBlockValidityProverBlockStatusByBlockHash
	err = s.dbApp.Exec(spanCtx, nil, func(d interface{}, _ interface{}) (err error) {
		q, _ := d.(SQLDriverApp)

		info, err = ucBlockValidityProverBlockStatusByBlockHash.New(s.config, s.log, q).Do(spanCtx, &input)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		switch {
		case errors.Is(err, errorsDB.ErrNotFound):
			return nil, utils.NotFound(
				spanCtx,
				fmt.Errorf("%s", block_validity_prover_block_status_by_block_hash.NotFoundMessage),
			)
		default:
			const msg = "failed to get block status by block hash: %+v"
			return &resp, utils.Internal(spanCtx, s.log, msg, err)
		}
	}

	resp.Success = true
	resp.Data = &node.BlockHashStatusResponse_Data{
		BlockNumber:                 uint32(info.BlockNumber),
		BlockHash:                   info.BlockHash,
		Status:                      node.BlockStatus_EXECUTED_ON_SCROLL,
		ExecutedBlockHashOnScroll:   info.ExecutedBlockHashOnScroll,
		ExecutedBlockHashOnEthereum: strings.TrimSpace(info.ExecutedBlockHashOnEthereum),
	}

	if !strings.EqualFold(resp.Data.ExecutedBlockHashOnEthereum, emptyKey) {
		resp.Data.Status = node.BlockStatus_EXECUTED_ON_ETHEREUM
	}

	return &resp, utils.OK(spanCtx)
}
