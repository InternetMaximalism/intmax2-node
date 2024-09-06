package scroll_eth

import (
	"context"
	"math/big"
)

type ScrollEth interface {
	GasFee(ctx context.Context) (*big.Int, error)
}
