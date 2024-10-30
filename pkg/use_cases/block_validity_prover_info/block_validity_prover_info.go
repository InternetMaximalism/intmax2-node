package block_validity_prover_info

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverInfo "intmax2-node/internal/use_cases/block_validity_prover_info"
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
) ucBlockValidityProverInfo.UseCaseBlockValidityProverInfo {
	return &uc{
		cfg: cfg,
		log: log,
		bvs: bvs,
	}
}

func (u *uc) Do(
	ctx context.Context,
) (*ucBlockValidityProverInfo.UCBlockValidityProverInfo, error) {
	const (
		hName = "UseCase BlockValidityProverInfo"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	info, err := u.bvs.FetchValidityProverInfo()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrFetchValidityProverInfoFail, err)
	}

	resp := ucBlockValidityProverInfo.UCBlockValidityProverInfo{
		DepositIndex: int64(info.DepositIndex),
		BlockNumber:  int64(info.BlockNumber),
	}

	return &resp, nil
}
