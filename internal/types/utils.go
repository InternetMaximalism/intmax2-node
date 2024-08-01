package types

import "math/big"

func BigIntToBytes32BeArray(bi *big.Int) [int32Key]byte {
	const int32Key = 32
	biBytes := bi.Bytes()
	var result [int32Key]byte
	copy(result[int32Key-len(biBytes):], biBytes)
	return result
}
