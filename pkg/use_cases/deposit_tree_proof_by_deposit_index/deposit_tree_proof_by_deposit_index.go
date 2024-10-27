package deposit_tree_proof_by_deposit_index

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucDepositTreeProofByDepositIndex "intmax2-node/internal/use_cases/deposit_tree_proof_by_deposit_index"

	"github.com/ethereum/go-ethereum/common"
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
) ucDepositTreeProofByDepositIndex.UseCaseDepositTreeProofByDepositIndex {
	return &uc{
		cfg: cfg,
		log: log,
		bvs: bvs,
	}
}

func (u *uc) Do(
	ctx context.Context, input *ucDepositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput,
) (*ucDepositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex, error) {
	const (
		hName           = "UseCase DepositTreeProofByDepositIndex"
		depositIndexKey = "deposit_index"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.Int64(depositIndexKey, input.DepositIndex),
		))
	defer span.End()

	depositMerkleProof, depositTreeRoot, err := u.bvs.DepositTreeProof(uint32(input.DepositIndex))
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrDepositTreeProofFail, err)
	}

	resp := ucDepositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex{
		MerkleProof: &ucDepositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexMerkleProof{
			Siblings: make([]string, len(depositMerkleProof.Siblings)),
		},
		RootHash: depositTreeRoot.String(),
	}

	for index := range depositMerkleProof.Siblings {
		resp.MerkleProof.Siblings[index] = common.BytesToHash(depositMerkleProof.Siblings[index][:]).String()
	}

	return &resp, nil
}