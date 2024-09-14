package types

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"intmax2-node/internal/finite_field"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
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

func (b *Bytes32) Equal(other *Bytes32) bool {
	for i := 0; i < int32Key/numUint32Bytes; i++ {
		if b[i] != other[i] {
			return false
		}
	}

	return true
}

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

func (b *Bytes32) ToFieldElementSlice() []ffg.Element {
	buf := make([]ffg.Element, 0)
	for _, x := range b {
		buf = append(buf, *new(ffg.Element).SetUint64(uint64(x)))
	}

	return buf
}

func (b *Bytes32) FromPoseidonHashOut(value *PoseidonHashOut) *Bytes32 {
	for i, e := range value.Elements {
		rawValue := e.ToUint64Regular()
		low := uint32(rawValue)
		high := uint32(rawValue >> 32)

		b[i*2] = high
		b[i*2+1] = low
	}

	return b
}

func (b *Bytes32) PoseidonHashOut() *PoseidonHashOut {
	elements := [4]ffg.Element{}
	for i := 0; i < len(elements); i++ {
		value := uint64(b[i*2])<<32 + uint64(b[i*2+1])
		elements[i].SetUint64(value)
	}

	return &PoseidonHashOut{Elements: elements}
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
func (v *Uint256) FromBigInt(value *big.Int) *Uint256 {
	copied_value := new(big.Int).Set(value)

	for i := 0; i < int8Key; i++ {
		v.inner[int8Key-1-i] = uint32(copied_value.Uint64())
		copied_value.Rsh(copied_value, int32Key)
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

func (v *Uint256) FromFieldElementSlice(value []ffg.Element) *Uint256 {
	for i, x := range value {
		y := x.ToUint64Regular()
		if y >= uint64(1)<<int32Key {
			panic("overflow")
		}
		v.inner[i] = uint32(y)
	}

	return v
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

func (v *Uint256) Bytes() []byte {
	return Uint32SliceToBytes(v.inner[:])
}

func (v *Uint256) FromBytes(bytes []byte) *Uint256 {
	for i := 0; i < int8Key; i++ {
		v.inner[i] = binary.BigEndian.Uint32(bytes[i*int4Key : (i+1)*int4Key])
	}

	return v
}

func (v *Uint256) Add(other *Uint256) *Uint256 {
	result := new(big.Int).Add(v.BigInt(), other.BigInt())
	return new(Uint256).FromBigInt(result)
}

func (v *Uint256) Sub(other *Uint256) *Uint256 {
	result := new(big.Int).Sub(v.BigInt(), other.BigInt())

	if result.Cmp(big.NewInt(0)) < 0 {
		panic("U256 sub underflow occurred")
	}

	return new(Uint256).FromBigInt(result)
}

type FlatG1 = [2]Bytes32

func FlattenG1Affine(pk *bn254.G1Affine) FlatG1 {
	x := Bytes32{}
	y := Bytes32{}
	pkX := pk.X.Bytes()
	pkY := pk.Y.Bytes()
	x.FromBytes(pkX[:])
	y.FromBytes(pkY[:])

	return [2]Bytes32{x, y}
}

func NewG1AffineFromFlatG1(v *FlatG1) *bn254.G1Affine {
	p := new(bn254.G1Affine)
	p.X.SetBytes(v[0].Bytes())
	p.Y.SetBytes(v[1].Bytes())

	return p
}

type FlatG2 = [int4Key]Bytes32

func FlattenG2Affine(sig *bn254.G2Affine) FlatG2 {
	xA0 := Bytes32{}
	xA1 := Bytes32{}
	yA0 := Bytes32{}
	yA1 := Bytes32{}
	pkXA0 := sig.X.A0.Bytes()
	pkXA1 := sig.X.A1.Bytes()
	pkYA0 := sig.Y.A0.Bytes()
	pkYA1 := sig.Y.A1.Bytes()
	xA0.FromBytes(pkXA0[:])
	xA1.FromBytes(pkXA1[:])
	yA0.FromBytes(pkYA0[:])
	yA1.FromBytes(pkYA1[:])

	return [int4Key]Bytes32{xA1, xA0, yA1, yA0}
}

func NewG2AffineFromFlatG2(v *FlatG2) *bn254.G2Affine {
	// x_a1, x_a0, y_a1, y_a0
	p := new(bn254.G2Affine)
	p.X.A1.SetBytes(v[0].Bytes())
	p.X.A0.SetBytes(v[1].Bytes())
	p.Y.A1.SetBytes(v[2].Bytes())
	p.Y.A0.SetBytes(v[3].Bytes())

	return p
}
