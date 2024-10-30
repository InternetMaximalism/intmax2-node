package get_intmax_block_info

import (
	"context"
	"encoding/json"
	"intmax2-node/configs"
	service "intmax2-node/internal/intmax_block_service"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	getINTMAXBlockInfo "intmax2-node/internal/use_cases/get_intmax_block_info"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// uc describes use case
type uc struct {
	cfg *configs.Config
	log logger.Logger
}

func New(
	cfg *configs.Config,
	log logger.Logger,
) getINTMAXBlockInfo.UseCaseGetINTMAXBlockInfo {
	return &uc{
		cfg: cfg,
		log: log,
	}
}

func (u *uc) Do(ctx context.Context, args []string, blockHash string, blockNumber uint64) (resp json.RawMessage, err error) {
	const (
		hName          = "UseCase GetINTMAXBlockInfo"
		blockHashKey   = "block_hash"
		blockNumberKey = "block_number"
		emptyKey       = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName,
		trace.WithAttributes(
			attribute.String(blockHashKey, blockHash),
			attribute.Int64(blockNumberKey, int64(blockNumber)),
		))
	defer span.End()

	blockHash = strings.TrimSpace(blockHash)
	if !strings.EqualFold(blockHash, emptyKey) {
		return service.INTMAXBlockInfoByBlockHash(spanCtx, u.cfg, blockHash)
	}

	return service.INTMAXBlockInfoByBlockNumber(spanCtx, u.cfg, blockNumber)
}
