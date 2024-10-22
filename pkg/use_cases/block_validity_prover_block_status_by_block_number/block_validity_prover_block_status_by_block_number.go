package block_validity_prover_block_status_by_block_number

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverBlockStatusByBlockNumber "intmax2-node/internal/use_cases/block_validity_prover_block_status_by_block_number"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type uc struct {
	cfg *configs.Config
	log logger.Logger
	db  SQLDriverApp
}

func New(
	cfg *configs.Config,
	log logger.Logger,
	db SQLDriverApp,
) ucBlockValidityProverBlockStatusByBlockNumber.UseCaseBlockValidityProverBlockStatusByBlockNumber {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *ucBlockValidityProverBlockStatusByBlockNumber.UCBlockValidityProverBlockStatusByBlockNumberInput,
) (*ucBlockValidityProverBlockStatusByBlockNumber.UCBlockValidityProverBlockStatusByBlockNumber, error) {
	const (
		hName          = "UseCase BlockValidityProverBlockStatusByBlockNumber"
		blockNumberKey = "block_number"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.Int64(blockNumberKey, int64(input.BlockNumber)),
		))
	defer span.End()

	bc, err := u.db.BlockContentByBlockNumber(uint32(input.BlockNumber))
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrBlockContentByBlockNumberFail, err)
	}

	const key0x = "0x"
	bc.BlockHash = common.HexToHash(fmt.Sprintf("%s%s", key0x, strings.TrimLeft(bc.BlockHash, key0x))).String()

	info := ucBlockValidityProverBlockStatusByBlockNumber.UCBlockValidityProverBlockStatusByBlockNumber{
		BlockNumber:               int64(bc.BlockNumber),
		BlockHash:                 bc.BlockHash,
		ExecutedBlockHashOnScroll: bc.BlockHashL2,
	}

	var relBcL2Bi *mDBApp.RelationshipL2BatchIndexBlockContents
	relBcL2Bi, err = u.db.RelationshipL2BatchIndexAndBlockContentsByBlockContentID(bc.BlockContentID)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrRelationshipL2BatchIndexAndBlockContentsByBlockContentIDFail, err)
	} else if errors.Is(err, errorsDB.ErrNotFound) {
		return &info, nil
	}

	var bi *mDBApp.L2BatchIndex
	bi, err = u.db.L2BatchIndex(&relBcL2Bi.L2BatchIndex)
	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrL2BatchIndexFail, err)
	} else if errors.Is(err, errorsDB.ErrNotFound) {
		return &info, nil
	}

	info.ExecutedBlockHashOnEthereum = bi.L1VerifiedBatchTxHash

	return &info, nil
}
