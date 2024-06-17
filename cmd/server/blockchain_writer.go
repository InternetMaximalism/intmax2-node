package server

import (
	"context"
)

type WriteBlockchain interface {
	RollupWB
}

type RollupWB interface {
	UpdateBlockBuilder(
		ctx context.Context,
		url string,
	) error
}
