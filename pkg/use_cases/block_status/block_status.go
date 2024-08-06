package block_status

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockStatus "intmax2-node/internal/use_cases/block_status"
	"intmax2-node/internal/worker"
	"strconv"

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
		return nil, err
	}

	isPosted := false
	var blockNumber string = "0"
	if *block.Status == 1 && block.BlockNumber != nil {
		isPosted = true
		blockNumber = strconv.FormatInt(*block.BlockNumber, 10)
	}

	status = &ucBlockStatus.UCBlockStatus{
		IsPosted:    isPosted,
		BlockNumber: blockNumber,
	}

	return status, nil
}

var ErrTransactionHashNotFound = errors.New("transaction hash not found")
var ErrTxTreeNotBuild = errors.New("tx tree not build")
var ErrTxTreeSignatureCollectionComplete = errors.New("tx tree signature collection complete")
var ErrValueInvalid = errors.New("value invalid")

func ExistsTxHash(w Worker, txHash string) (txTree *worker.TxTree, err error) {
	info, err := w.TrHash(txHash)
	if err != nil && errors.Is(err, worker.ErrTransactionHashNotFound) {
		return nil, ErrTransactionHashNotFound
	}
	fmt.Printf("ExistsTxHash txHash: %s", txHash)
	info.TxHash = txHash

	txTree, err = w.TxTreeByAvailableFile(info)
	if err != nil {
		switch {
		case errors.Is(err, worker.ErrTxTreeByAvailableFileFail):
			return nil, ErrTransactionHashNotFound
		case errors.Is(err, worker.ErrTxTreeNotFound):
			return nil, ErrTxTreeNotBuild
		case errors.Is(err, worker.ErrTxTreeSignatureCollectionComplete):
			return nil, ErrTxTreeSignatureCollectionComplete
		default:
			return nil, ErrValueInvalid
		}
	}

	return txTree, nil
}
