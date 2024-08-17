package types

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"intmax2-node/internal/finite_field"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const numUint32Bytes = 4

func BigIntToBytes32BeArray(bi *big.Int) [int32Key]byte {
	biBytes := bi.Bytes()
	var result [int32Key]byte
	copy(result[int32Key-len(biBytes):], biBytes)
	return result
}

type Bytes16 [int16Key / numUint32Bytes]uint32

func (b *Bytes16) FromBytes(bytes []byte) {
	for i := 0; i < int16Key/numUint32Bytes; i++ {
		b[i] = binary.BigEndian.Uint32(bytes[i*numUint32Bytes : (i+1)*numUint32Bytes])
	}
}

func (b *Bytes16) Bytes() []byte {
	bytes := make([]byte, int16Key)
	for i := 0; i < int16Key/numUint32Bytes; i++ {
		binary.BigEndian.PutUint32(bytes[i*numUint32Bytes:(i+1)*numUint32Bytes], b[i])
	}

	return bytes
}

func (b *Bytes16) Hex() string {
	return hexutil.Encode(b.Bytes())
}

func (b *Bytes16) FromHex(s string) error {
	bytes, err := hexutil.Decode(s)
	if err != nil {
		return err
	}

	b.FromBytes(bytes)
	return nil
}

func (b *Bytes16) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Hex())
}

func (b *Bytes16) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	return b.FromHex(s)
}

type Bytes32 [int32Key / numUint32Bytes]uint32

func (b *Bytes32) FromBytes(bytes []byte) {
	for i := 0; i < int32Key/numUint32Bytes; i++ {
		b[i] = binary.BigEndian.Uint32(bytes[i*numUint32Bytes : (i+1)*numUint32Bytes])
	}
}

func (b *Bytes32) Bytes() []byte {
	bytes := make([]byte, int32Key)
	for i := 0; i < int32Key/numUint32Bytes; i++ {
		binary.BigEndian.PutUint32(bytes[i*numUint32Bytes:(i+1)*numUint32Bytes], b[i])
	}

	return bytes
}

func (b *Bytes32) Hex() string {
	return hexutil.Encode(b.Bytes())
}

func (b *Bytes32) FromHex(s string) error {
	bytes, err := hexutil.Decode(s)
	if err != nil {
		return err
	}

	b.FromBytes(bytes)
	return nil
}

func (b *Bytes32) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Hex())
}

func (b *Bytes32) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	return b.FromHex(s)
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

func (v *Uint256) BigInt() *big.Int {
	res := big.NewInt(0)
	for i := 0; i < int8Key; i++ {
		res.Lsh(res, int32Key)
		res.Add(res, big.NewInt(int64(v.inner[i])))
	}

	return res
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

// convert 256 bits number
func (v *Uint256) MarshalJSON() ([]byte, error) {
	s := v.BigInt().String()
	return json.Marshal(s)
}

func (v *Uint256) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bi, ok := new(big.Int).SetString(s, int10Key)
	if !ok {
		var ErrNotNumberString = errors.New("not a number string")
		return ErrNotNumberString
	}

	v.FromBigInt(bi)
	return nil
}

type FlatG1 = [2]Uint256

type FlatG2 = [int4Key]Uint256
