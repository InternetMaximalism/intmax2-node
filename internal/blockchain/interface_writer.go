package blockchain

import (
	"context"
)

type WriteBlockchain interface {
	RollupW
}

type RollupW interface {
	UpdateBlockBuilder(
		ctx context.Context,
		url string,
	) error
	StopBlockBuilder(
		ctx context.Context,
	) error
}
