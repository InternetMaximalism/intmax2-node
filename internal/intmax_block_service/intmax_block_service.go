package intmax_block_service

import (
	"context"
	"encoding/json"
	"intmax2-node/configs"
)

func INTMAXBlockInfoByBlockHash(
	ctx context.Context,
	cfg *configs.Config,
	blockHash string,
) (json.RawMessage, error) {
	return GetINTMAXBlockInfoByHashWithRawRequest(ctx, cfg, blockHash)
}

func INTMAXBlockInfoByBlockNumber(
	ctx context.Context,
	cfg *configs.Config,
	blockNumber uint64,
) (json.RawMessage, error) {
	return GetINTMAXBlockInfoByNumberWithRawRequest(ctx, cfg, blockNumber)
}
