package types

import (
	"encoding/binary"
	"intmax2-node/internal/finite_field"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

func BigIntToBytes32BeArray(bi *big.Int) [int32Key]byte {
	biBytes := bi.Bytes()
	var result [int32Key]byte
	copy(result[int32Key-len(biBytes):], biBytes)
	return result
}

type Bytes16 [int4Key]uint32

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
	buf := make([]byte, len(v)*int4Key)
	for i, n := range v {
		binary.BigEndian.PutUint32(buf[i*int4Key:], n)
	}

	return buf
}

type Uint256 struct {
	inner [int8Key]uint32
}

func (v *Uint256) Equal(other *Uint256) bool {
	for i := 0; i < int8Key; i++ {
		if v.inner[i] != other.inner[i] {
			return false
		}
	}

	return true
}

// FromBigInt converts a big.Int to a Uint256
// the big.Int is split into 8 32-bit words (big-endian)
func (v *Uint256) FromBigInt(a *big.Int) *Uint256 {
	for i := 0; i < int8Key; i++ {
		v.inner[int8Key-1-i] = uint32(a.Uint64())
		a.Rsh(a, int32Key)
	}

	return v
}

func (v *Uint256) ToFieldElementSlice() []ffg.Element {
	res := finite_field.NewBuffer(make([]ffg.Element, 0))
	for i := 0; i < int8Key; i++ {
		err := finite_field.WriteUint64(res, uint64(v.inner[i]))
		if err != nil {
			panic(err)
		}
	}

	return res.Inner()
}

func (v *Uint256) IsDummyPublicKey() bool {
	one := new(Uint256).FromBigInt(big.NewInt(1))
	return v.Equal(one)
}

type FlatG1 = [2]Uint256

type FlatG2 = [int4Key]Uint256
