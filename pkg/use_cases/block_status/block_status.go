package block_status

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockStatus "intmax2-node/internal/use_cases/block_status"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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
	worker Worker,
) ucBlockStatus.UseCaseBlockStatus {
	return &uc{
		cfg:    cfg,
		log:    log,
		db:     db,
		worker: worker,
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
		fmt.Printf("BlockByTxRoot error\n")
		if err.Error() == "not found" {
			err = u.worker.ExistsTxTreeRoot(input.TxTreeRoot)
			if err != nil {
				return nil, err
			}

			status = &ucBlockStatus.UCBlockStatus{
				IsPosted:    false,
				BlockNumber: 0,
			}

			return status, nil
		}

		return nil, err
	}
	fmt.Printf("block: %v\n", block)

	isPosted := false
	var blockNumber uint32 = 0
	if *block.Status == 1 && block.BlockNumber != nil {
		isPosted = true
		blockNumber = uint32(*block.BlockNumber)
	}

	status = &ucBlockStatus.UCBlockStatus{
		IsPosted:    isPosted,
		BlockNumber: blockNumber,
	}

	return status, nil
}
