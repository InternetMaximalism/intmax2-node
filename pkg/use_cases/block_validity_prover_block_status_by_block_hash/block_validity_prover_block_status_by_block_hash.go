package block_validity_prover_block_status_by_block_hash

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	ucBlockValidityProverBlockStatusByBlockHash "intmax2-node/internal/use_cases/block_validity_prover_block_status_by_block_hash"
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
) ucBlockValidityProverBlockStatusByBlockHash.UseCaseBlockValidityProverBlockStatusByBlockHash {
	return &uc{
		cfg: cfg,
		log: log,
		db:  db,
	}
}

func (u *uc) Do(
	ctx context.Context,
	input *ucBlockValidityProverBlockStatusByBlockHash.UCBlockValidityProverBlockStatusByBlockHashInput,
) (*ucBlockValidityProverBlockStatusByBlockHash.UCBlockValidityProverBlockStatusByBlockHash, error) {
	const (
		hName        = "UseCase BlockValidityProverBlockStatusByBlockHash"
		blockHashKey = "block_hash"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(blockHashKey, input.BlockHash),
		))
	defer span.End()

	const key0x = "0x"
	input.BlockHash = fmt.Sprintf("%s%s", key0x, strings.TrimLeft(input.BlockHash, key0x))
	input.BlockHash = strings.TrimLeft(common.HexToHash(input.BlockHash).Hex(), key0x)

	bc, err := u.db.BlockContentByBlockHash(input.BlockHash)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrBlockContentByBlockHashFail, err)
	}

	bc.BlockHash = common.HexToHash(fmt.Sprintf("%s%s", key0x, strings.TrimLeft(bc.BlockHash, key0x))).String()

	info := ucBlockValidityProverBlockStatusByBlockHash.UCBlockValidityProverBlockStatusByBlockHash{
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
