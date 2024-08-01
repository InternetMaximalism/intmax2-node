package block_post_service

import (
	"context"
	"intmax2-node/internal/bindings"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BlockPostService interface {
	FetchLatestBlockNumber(ctx context.Context) (uint64, error)
	FetchNewPostedBlocks(startBlock uint64) ([]*bindings.RollupBlockPosted, *big.Int, error)
	FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error)
}
