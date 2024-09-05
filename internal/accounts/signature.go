package accounts

import (
	"encoding/binary"
	"errors"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const int4Key = 4

// Sign is calculate the signature of the message using the private key as follows:
// messagePoint = hashToG2(message)
// signature = privateKey * messagePoint
func (a PrivateKey) Sign(message []ffg.Element) (*bn254.G2Affine, error) {
	if a.sk == nil {
		err := errors.New("private key is nil")
		return nil, err
	}

	messagePoint := goldenposeidon.HashToG2(message)

	// Multiply the private key with the message point
	// signature := new(bn256.G2).ScalarMult(messagePoint, a.sk)
	signature := new(bn254.G2Affine).ScalarMultiplication(&messagePoint, a.sk)

	return signature, nil
}

// VerifySignature is verify the signature of the message using the public key as follows:
// e(publicKey, hashToG2(message)) = e(G1, signature), where G1 is the generator of the group
func VerifySignature(signature *bn254.G2Affine, publicKey *PublicKey, message []ffg.Element) error {
	messagePoint := goldenposeidon.HashToG2(message)

	g1Generator := new(bn254.G1Affine).ScalarMultiplicationBase(big.NewInt(1))
	g1GeneratorInv := new(bn254.G1Affine).Neg(g1Generator)

	g1s := []bn254.G1Affine{*g1GeneratorInv, *publicKey.Pk}
	g2s := []bn254.G2Affine{*signature, messagePoint}

	success, err := bn254.PairingCheck(g1s, g2s)

	if err != nil {
		return err
	}

	if !success {
		return errors.New("signature verification failed")
	}

	return nil
}

// WeightByHash calculates a weighted private key based on the provided hash.
// It first computes a coefficient using the hash and the public key, then
// returns a new PrivateKey instance with the public key scaled by this coefficient.
func (pk *PublicKey) WeightByHash(hash []byte) *PublicKey {
	coeff := hashToWeight(&pk.Pk.X, hash)
	return pk.weight(coeff)
}

// weight applies the given coefficient to the public key,
// returning a new weighted PublicKey instance.
func (pk *PublicKey) weight(coeff *big.Int) *PublicKey {
	weighted := new(PublicKey)
	weighted.Pk = new(bn254.G1Affine).ScalarMultiplication(pk.Pk, coeff)

	return weighted
}

// WeightByHash calculates a weighted private key based on the provided hash.
// It first computes a coefficient using the hash and the public key, then
// returns a new PrivateKey instance with the private key and public key scaled by this coefficient.
func (a *PrivateKey) WeightByHash(publicKeysHash []byte) *PrivateKey {
	coeff := hashToWeight(&a.Pk.X, publicKeysHash)
	return a.weight(coeff)
}

// weight applies the given coefficient to the private key and public key,
// returning a new weighted PrivateKey instance.
func (a *PrivateKey) weight(coeff *big.Int) *PrivateKey {
	weighted := new(PrivateKey)
	weighted.sk = new(big.Int).Mul(a.sk, coeff)
	weighted.PublicKey = *a.Public().weight(coeff)

	return weighted
}

// hashToWeight calculates a weighting factor (coefficient) based on the public key and hash.
// It combines the public key and hash into a byte slice, converts it to field elements,
// and derive the weight.
func hashToWeight(myPublicKey *fp.Element, hash []byte) *big.Int {
	p := myPublicKey.Bytes()
	flatten := []byte{}
	flatten = append(flatten, p[:]...)
	flatten = append(flatten, hash...)

	flatten2 := make([]ffg.Element, len(flatten)/int4Key)
	for i := 0; i < len(flatten)/int4Key; i++ {
		v := binary.BigEndian.Uint32(flatten[i*int4Key : (i+1)*int4Key])
		flatten2[i].SetUint64(uint64(v))
	}
	challenger := goldenposeidon.NewChallenger()
	challenger.ObserveElements(flatten2)

	const (
		bn254OrderByteSize = 32
		uint32ByteSize     = 4
	)
	output := challenger.GetNChallenges(bn254OrderByteSize / uint32ByteSize)

	return goldenposeidon.FieldElementSliceToBigInt(output)
}
