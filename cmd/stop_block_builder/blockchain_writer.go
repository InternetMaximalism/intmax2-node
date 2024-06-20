package stop_block_builder

import (
	"context"
)

type WriteBlockchain interface {
	RollupWB
}

type RollupWB interface {
	StopBlockBuilder(
		ctx context.Context,
	) error
}
