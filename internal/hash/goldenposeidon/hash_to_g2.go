package goldenposeidon

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/iden3/go-iden3-crypto/ffg"
)

func FieldElementSliceToBigInt(inputs []*ffg.Element) *big.Int {
	value := big.NewInt(0)
	const shift = 32
	base := new(big.Int).SetUint64(uint64(1) << shift)
	power := big.NewInt(1)
	for _, c := range inputs {
		limb := uint32(c.ToUint64Regular())
		x := new(big.Int).SetUint64(uint64(limb))

		// Calculate value += limb * power
		x.Mul(x, power)
		value.Add(value, x)

		// power *= base
		power.Mul(power, base)
	}

	return value
}

func HashToFq2(inputs []ffg.Element) bn254.E2 {
	challenger := NewChallenger()
	challenger.ObserveElements(inputs)

	const nChallenges = 2 * 8
	c0Output := challenger.GetNChallenges(nChallenges)
	c0 := FieldElementSliceToBigInt(c0Output)

	c1Output := challenger.GetNChallenges(nChallenges)
	c1 := FieldElementSliceToBigInt(c1Output)

	return bn254.E2{
		A0: *new(fp.Element).SetBigInt(c0),
		A1: *new(fp.Element).SetBigInt(c1),
	}
}

// nolint:gocritic
func ClearCofactor(a bn254.G2Affine) bn254.G2Affine {
	const radix = 10
	cofactor, ok := new(big.Int).SetString("21888242871839275222246405745257275088844257914179612981679871602714643921549", radix)
	if !ok {
		panic("failed to parse cofactor")
	}

	// Calculate result := b1 * cofactor
	result := bn254.G2Affine{
		X: bn254.E2{
			A0: fp.Element{0},
			A1: fp.Element{0},
		},
		Y: bn254.E2{
			A0: fp.Element{0},
			A1: fp.Element{0},
		},
	}
	x := cofactor
	for i := x.BitLen() - 1; i >= 0; i-- {
		result.Double(&result)
		if x.Bit(i) == 1 {
			result.Add(&result, &a)
		}
	}

	return result
}

func MapToG2(u bn254.E2) bn254.G2Affine {
	a := bn254.MapToCurve2(&u)

	return ClearCofactor(a)
}

func HashToG2(inputs []ffg.Element) bn254.G2Affine {
	u := HashToFq2(inputs)
	return MapToG2(u)
}
