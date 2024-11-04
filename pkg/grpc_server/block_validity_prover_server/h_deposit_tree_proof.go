package block_validity_prover_server

import (
	"context"
	"errors"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	depositTreeProofByDepositIndex "intmax2-node/internal/use_cases/deposit_tree_proof_by_deposit_index"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) DepositTreeProof(
	ctx context.Context,
	req *node.DepositTreeProofRequest,
) (*node.DepositTreeProofResponse, error) {
	resp := node.DepositTreeProofResponse{}

	const (
		hName      = "Handler DepositTreeProof"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput{
		DepositIndex: req.DepositIndex,
		BlockNumber:  req.BlockNumber,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var info *depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex
	info, err = s.Commands().DepositTreeProofByDepositIndex(s.config, s.log, s.bvs).Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		if errors.Is(err, block_validity_prover.ErrBlockNumberInvalid) ||
			errors.Is(err, block_validity_prover.ErrBlockNumberOutOfRange) {
			return &resp, utils.NotFound(spanCtx, block_validity_prover.ErrNoValidityProofByBlockNumber)
		}

		const msg = "failed to get deposit tree proof by deposit index: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.DepositTreeProofResponse_Data{
		MerkleProof: &node.DepositTreeProofResponse_MerkleProof{
			Siblings: make([]string, len(info.MerkleProof.Siblings)),
		},
		RootHash: info.RootHash,
	}
	copy(resp.Data.MerkleProof.Siblings, info.MerkleProof.Siblings)

	return &resp, utils.OK(spanCtx)
}
