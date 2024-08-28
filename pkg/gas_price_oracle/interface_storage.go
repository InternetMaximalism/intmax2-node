package gas_price_oracle

import (
	"context"
	"math/big"
)

type Storage interface {
	Init(ctx context.Context) (err error)
	Value(ctx context.Context, name string) (*big.Int, error)
	UpdValue(ctx context.Context, name string) (err error)
	UpdValues(ctx context.Context, name ...string) (err error)
}
