package types

import (
	"encoding/binary"
	"math/big"
)

func BigIntToBytes32BeArray(bi *big.Int) [int32Key]byte {
	biBytes := bi.Bytes()
	var result [int32Key]byte
	copy(result[int32Key-len(biBytes):], biBytes)
	return result
}

type Bytes32 [int8Key]uint32

func (b *Bytes32) FromBytes(bytes []byte) {
	for i := 0; i < int8Key; i++ {
		b[i] = binary.BigEndian.Uint32(bytes[i*int4Key : (i+1)*int4Key])
	}
}

func (b *Bytes32) Bytes() []byte {
	bytes := make([]byte, int8Key*int4Key)
	for i := 0; i < int8Key; i++ {
		binary.BigEndian.PutUint32(bytes[i*int4Key:(i+1)*int4Key], b[i])
	}

	return bytes
}

func Uint32SliceToBytes(v []uint32) []byte {
	const int4Key = 4

	buf := make([]byte, len(v)*int4Key)
	for i, n := range v {
		binary.BigEndian.PutUint32(buf[i*int4Key:], n)
	}

	return buf
}
