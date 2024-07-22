package types

import "math/big"

type Balance struct {
	TokenIndex uint32
	Amount     *big.Int
}
