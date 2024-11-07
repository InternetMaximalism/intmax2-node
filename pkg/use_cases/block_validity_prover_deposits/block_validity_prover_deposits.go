package block_validity_prover_deposits

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverDeposits "intmax2-node/internal/use_cases/block_validity_prover_deposits"

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
) ucBlockValidityProverDeposits.UseCaseBlockValidityProverDeposits {
	return &uc{
		cfg: cfg,
		log: log,
		bvs: bvs,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *ucBlockValidityProverDeposits.UCBlockValidityProverDepositsInput,
) ([]*ucBlockValidityProverDeposits.UCBlockValidityProverDeposits, error) {
	const (
		hName            = "UseCase BlockValidityProverDeposits"
		depositHashesKey = "deposit_hashes"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	if input == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrUCBlockValidityProverDepositsInputEmpty)
		return nil, ErrUCBlockValidityProverDepositsInputEmpty
	}

	span.SetAttributes(
		attribute.StringSlice(depositHashesKey, input.DepositHashes),
	)

	list, err := u.bvs.GetDepositsInfoByHash(input.ConvertDepositHashes...)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, err
	}

	var index int
	result := make([]*ucBlockValidityProverDeposits.UCBlockValidityProverDeposits, len(list))
	for key := range list {
		result[index] = &ucBlockValidityProverDeposits.UCBlockValidityProverDeposits{
			DepositId:      list[key].DepositId,
			DepositHash:    list[key].DepositHash,
			DepositIndex:   list[key].DepositIndex,
			BlockNumber:    list[key].BlockNumber,
			IsSynchronized: list[key].IsSynchronized,
			DepositLeaf:    list[key].DepositLeaf,
			Sender:         list[key].Sender,
		}
		index++
	}

	return result, nil
}
