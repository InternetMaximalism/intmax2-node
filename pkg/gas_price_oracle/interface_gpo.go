package gas_price_oracle

import (
	"context"
	"math/big"
)

type GasPriceOracle interface {
	// GasFee returns gasFee in wei
	GasFee(ctx context.Context) (*big.Int, error)
}
