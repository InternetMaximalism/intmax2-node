package block_tree_proof_by_root_and_leaf_block_numbers

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockTreeProofByRootAndLeafBlockNumbers "intmax2-node/internal/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type uc struct {
	cfg *configs.Config
	log logger.Logger
	bvs BlockValidityService
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	bvs BlockValidityService,
) ucBlockTreeProofByRootAndLeafBlockNumbers.UseCaseBlockTreeProofByRootAndLeafBlockNumbers {
	return &uc{
		cfg: cfg,
		log: log,
		bvs: bvs,
	}
}

func (u *uc) Do(
	ctx context.Context, input *ucBlockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbersInput,
) (*ucBlockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbers, error) {
	const (
		hName              = "UseCase BlockTreeProofByRootAndLeafBlockNumbers"
		rootBlockNumberKey = "root_block_number"
		leafBlockNumberKey = "leaf_block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.Int64(rootBlockNumberKey, input.RootBlockNumber),
			attribute.Int64(leafBlockNumberKey, input.LeafBlockNumber),
		))
	defer span.End()

	proof, root, err := u.bvs.BlockTreeProof(uint32(input.RootBlockNumber), uint32(input.LeafBlockNumber))
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrBlockTreeProofFail, err)
	}

	resp := ucBlockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbers{
		MerkleProof: &ucBlockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbersMerkleProof{
			Siblings: make([]string, len(proof.Siblings)),
		},
		RootHash: root.String(),
	}

	for index := range proof.Siblings {
		resp.MerkleProof.Siblings[index] = proof.Siblings[index].String()
	}

	return &resp, nil
}
