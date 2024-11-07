package block_validity_prover_block_validity_public_inputs

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverBlockValidityPublicInputs "intmax2-node/internal/use_cases/block_validity_prover_block_validity_public_inputs"

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
) ucBlockValidityProverBlockValidityPublicInputs.UseCaseBlockValidityProverBlockValidityPublicInputs {
	return &uc{
		cfg: cfg,
		log: log,
		bvs: bvs,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *ucBlockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputsInput,
) (*ucBlockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs, error) {
	const (
		hName          = "UseCase BlockValidityProverBlockValidityPublicInputs"
		blockNumberKey = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCBlockValidityProverBlockValidityPublicInputsInputEmpty)
		return nil, ErrUCBlockValidityProverBlockValidityPublicInputsInputEmpty
	}

	span.SetAttributes(
		attribute.Int64(blockNumberKey, input.BlockNumber),
	)

	validityPublicInputs, senderLeaves, err := u.bvs.ValidityPublicInputsByBlockNumber(uint32(input.BlockNumber))
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, err
	}

	return &ucBlockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
		ValidityPublicInputs: validityPublicInputs,
		Sender:               senderLeaves,
	}, nil
}
