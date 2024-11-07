package block_validity_prover_block_validity_proof

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverBlockValidityProof "intmax2-node/internal/use_cases/block_validity_prover_block_validity_proof"

	"go.opentelemetry.io/otel/attribute"
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
) ucBlockValidityProverBlockValidityProof.UseCaseBlockValidityProverBlockValidityProof {
	return &uc{
		cfg: cfg,
		log: log,
		bvs: bvs,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *ucBlockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProofInput,
) (*ucBlockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProof, error) {
	const (
		hName          = "UseCase BlockValidityProverBlockValidityProof"
		blockNumberKey = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCBlockValidityProverBlockValidityProofInputEmpty)
		return nil, ErrUCBlockValidityProverBlockValidityProofInputEmpty
	}

	span.SetAttributes(
		attribute.Int64(blockNumberKey, input.BlockNumber),
	)

	validityProof, err := u.bvs.ValidityProofByBlockNumber(uint32(input.BlockNumber))
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, err
	}

	return &ucBlockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProof{
		ValidityPublicInputs: validityProof.ValidityPublicInputs,
		ValidityProof:        validityProof.ValidityProof,
		Sender:               validityProof.SenderLeaf,
	}, nil
}
