package block_validity_prover_server

import (
	"context"
	"intmax2-node/internal/open_telemetry"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	blockTreeProofByRootAndLeafBlockNumbers "intmax2-node/internal/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"
	"intmax2-node/pkg/grpc_server/utils"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *BlockValidityProverServer) BlockTreeProof(
	ctx context.Context,
	req *node.BlockTreeProofRequest,
) (*node.BlockTreeProofResponse, error) {
	resp := node.BlockTreeProofResponse{}

	const (
		hName      = "Handler BlockTreeProof"
		requestKey = "request"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(requestKey, req.String()),
		))
	defer span.End()

	input := blockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbersInput{
		RootBlockNumber: req.RootBlockNumber,
		LeafBlockNumber: req.LeafBlockNumber,
	}

	err := input.Valid()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return &resp, utils.BadRequest(spanCtx, err)
	}

	var info *blockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbers
	info, err = s.Commands().BlockTreeProofByRootAndLeafBlockNumbers(s.config, s.log, s.bvs).Do(spanCtx, &input)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		const msg = "failed to get block tree proof by root and lead block numbers: %+v"
		return &resp, utils.Internal(spanCtx, s.log, msg, err)
	}

	resp.Success = true
	resp.Data = &node.BlockTreeProofResponse_Data{
		MerkleProof: &node.BlockTreeProofResponse_MerkleProof{
			Siblings: make([]string, len(info.MerkleProof.Siblings)),
		},
		RootHash: info.RootHash,
	}
	copy(resp.Data.MerkleProof.Siblings, info.MerkleProof.Siblings)

	return &resp, utils.OK(spanCtx)
}
