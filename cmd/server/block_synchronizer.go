package server

import (
	"intmax2-node/internal/bindings"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_block_synchronizer.go -package=server -source=block_synchronizer.go

type BlockSynchronizer interface {
	FetchNewPostedBlocks(startBlock uint64, endBlock *uint64) ([]*bindings.RollupBlockPosted, *big.Int, error)
	FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error)
}
