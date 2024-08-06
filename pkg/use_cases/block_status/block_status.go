package block_status

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockStatus "intmax2-node/internal/use_cases/block_status"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const base10 = 10

type uc struct {
	cfg    *configs.Config
	log    logger.Logger
	db     SQLDriverApp
	worker Worker
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
	w Worker,
) ucBlockStatus.UseCaseBlockStatus {
	return &uc{
		cfg:    cfg,
		log:    log,
		db:     db,
		worker: w,
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

	block, err := u.db.BlockByTxRoot(input.TxTreeRoot)
	if err != nil {
		return nil, err
	}

	isPosted := false
	blockNumber := "0"
	if *block.Status == 1 && block.BlockNumber != nil {
		isPosted = true
		blockNumber = strconv.FormatInt(*block.BlockNumber, base10)
	}

	status = &ucBlockStatus.UCBlockStatus{
		IsPosted:    isPosted,
		BlockNumber: blockNumber,
	}

	return status, nil
}
