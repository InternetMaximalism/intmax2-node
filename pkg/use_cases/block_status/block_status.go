package block_signature

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockStatus "intmax2-node/internal/use_cases/block_status"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type uc struct {
	cfg *configs.Config
	log logger.Logger
}

func New(
	cfg *configs.Config,
	log logger.Logger,
) ucBlockStatus.UseCaseBlockStatus {
	return &uc{
		cfg: cfg,
		log: log,
	}
}

func (u *uc) Do(
	ctx context.Context, input *ucBlockStatus.UCBlockStatusInput,
) (status *ucBlockStatus.UCBlockStatus, err error) {
	const (
		hName         = "UseCase BlockStatus"
		txTreeRootKey = "tx_tree_root"
	)

	_, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(txTreeRootKey, input.TxTreeRoot),
		))
	defer span.End()

	status = &ucBlockStatus.UCBlockStatus{
		IsPosted:    true,
		BlockNumber: 1,
	}

	return status, nil
}
