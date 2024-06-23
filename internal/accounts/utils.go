package accounts

import (
	"encoding/hex"

	"github.com/consensys/gnark-crypto/ecc/bn254"
)

func EncodeG1CurvePoint(point *bn254.G1Affine) string {
	// p.X, p.Y
	bytes := point.Marshal()
	return hex.EncodeToString(bytes)
}

func DecodeG1CurvePoint(encoded string) (*bn254.G1Affine, error) {
	bytes, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	point := new(bn254.G1Affine)
	err = point.Unmarshal(bytes)
	if err != nil {
		return nil, err
	}

	return point, nil
}

func EncodeG2CurvePoint(point *bn254.G2Affine) string {
	// p.X.A1, p.X.A0, p.Y.A1, p.Y.A0
	bytes := point.Marshal()
	return hex.EncodeToString(bytes)
}

func DecodeG2CurvePoint(encoded string) (*bn254.G2Affine, error) {
	bytes, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	point := new(bn254.G2Affine)
	err = point.Unmarshal(bytes)
	if err != nil {
		return nil, err
	}

	return point, nil
}
