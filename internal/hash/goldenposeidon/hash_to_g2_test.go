package goldenposeidon

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
)

func TestFieldElementSliceToBigInt(t *testing.T) {
	t.Parallel()

	uintSlice := []uint64{1, 2}
	fieldElementSlice := []*ffg.Element{}
	for _, v := range uintSlice {
		fieldElementSlice = append(fieldElementSlice, ffg.NewElementFromUint64(v))
	}
	actual := FieldElementSliceToBigInt(fieldElementSlice)
	a, ok := new(big.Int).SetString("200000001", 16)
	assert.True(t, ok)
	expected := new(fp.Element).SetBigInt(a)
	assert.Equal(t, actual, expected)

	uintSlice = []uint64{1, 2, 3}
	fieldElementSlice = []*ffg.Element{}
	for _, v := range uintSlice {
		fieldElementSlice = append(fieldElementSlice, ffg.NewElementFromUint64(v))
	}
	actual = FieldElementSliceToBigInt(fieldElementSlice)
	a, ok = new(big.Int).SetString("55340232229718589441", 10)
	assert.True(t, ok)
	expected = new(fp.Element).SetBigInt(a)
	assert.Equal(t, expected, actual)

	uintSlice = []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	fieldElementSlice = []*ffg.Element{}
	for _, v := range uintSlice {
		fieldElementSlice = append(fieldElementSlice, ffg.NewElementFromUint64(v))
	}
	actual = FieldElementSliceToBigInt(fieldElementSlice)
	a, ok = new(big.Int).SetString("13381388375079396747702625564911137379397638293461898465501511720667989245944", 10)
	assert.True(t, ok)
	expected = new(fp.Element).SetBigInt(a)
	assert.Equal(t, expected, actual)

	uintSlice = []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	fieldElementSlice = []*ffg.Element{}
	for _, v := range uintSlice {
		fieldElementSlice = append(fieldElementSlice, ffg.NewElementFromUint64(v))
	}
	actual = FieldElementSliceToBigInt(fieldElementSlice)
	a, ok = new(big.Int).SetString("13267190979198767422712211918368298814746222286259698739437271243137716783425", 10)
	assert.True(t, ok)
	expected = new(fp.Element).SetBigInt(a)
	assert.Equal(t, expected, actual)
}

func TestHashToFq2(t *testing.T) {
	t.Parallel()

	uintSlice := []uint64{1, 2}
	fieldElementSlice := []*ffg.Element{}
	for _, v := range uintSlice {
		fieldElementSlice = append(fieldElementSlice, ffg.NewElementFromUint64(v))
	}
	actual := HashToFq2(fieldElementSlice)
	fmt.Printf("actual: (%v, %v)\n", actual.A0.Marshal(), actual.A1.Marshal())
}

func TestMapToG2(t *testing.T) {
	t.Parallel()

	a := new(bn254.E2).SetOne()

	b := MapToG2(*a)
	expected, ok := new(big.Int).SetString("12994054379246242930997819300517812813323481445446407404988774464198245037454", 10)
	assert.True(t, ok)
	assert.Equal(t, expected, b.X.A0.BigInt(new(big.Int)))
	expected, ok = new(big.Int).SetString("759429597763187778895178026133809638044527903897013273428819372671458813267", 10)
	assert.True(t, ok)
	assert.Equal(t, expected, b.X.A1.BigInt(new(big.Int)))
	expected, ok = new(big.Int).SetString("1221977559152545693799606645020244795833559010546921632617067432230144783701", 10)
	assert.True(t, ok)
	assert.Equal(t, expected, b.Y.A0.BigInt(new(big.Int)))
	expected, ok = new(big.Int).SetString("642364649382206664496976264454552182727890291422982090459343786761001117187", 10)
	assert.True(t, ok)
	assert.Equal(t, expected, b.Y.A1.BigInt(new(big.Int)))

}

func TestHashToG2(t *testing.T) {
	t.Parallel()

	uintSlice := []uint64{1, 2}
	fieldElementSlice := []*ffg.Element{}
	for _, v := range uintSlice {
		fieldElementSlice = append(fieldElementSlice, ffg.NewElementFromUint64(v))
	}
	actual := HashToG2(fieldElementSlice)
	fmt.Printf("actual: %v\n", actual)
}
