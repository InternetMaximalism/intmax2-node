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

// Calculate the signature of the message using the private key as follows:
// messagePoint = hashToG2(message)
// signature = privateKey * messagePoint
func (a PrivateKey) Sign(message []*ffg.Element) (*bn254.G2Affine, error) {
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

// Verify the signature of the message using the public key as follows:
// e(publicKey, hashToG2(message)) = e(G1, signature), where G1 is the generator of the group
func VerifySignature(signature *bn254.G2Affine, publicKey *PublicKey, message []*ffg.Element) error {
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

func (pk *PublicKey) WeightByHash(hash []byte) *PublicKey {
	coeff := hashToWeight(&pk.Pk.X, hash)
	return pk.weight(coeff)
}

func (pk *PublicKey) weight(coeff *big.Int) *PublicKey {
	weighted := new(PublicKey)
	weighted.Pk = new(bn254.G1Affine).ScalarMultiplication(pk.Pk, coeff)

	return weighted
}

func (a *PrivateKey) WeightByHash(publicKeysHash []byte) *PrivateKey {
	coeff := hashToWeight(&a.Pk.X, publicKeysHash)
	return a.weight(coeff)
}

func (a *PrivateKey) weight(coeff *big.Int) *PrivateKey {
	weighted := new(PrivateKey)
	weighted.sk = new(big.Int).Mul(a.sk, coeff)
	weighted.PublicKey = *a.Public().weight(coeff)

	return weighted
}

func HashToFieldElementSlice(hash []byte) ([]*ffg.Element, error) {
	const uint32ByteSize = 4
	hashByteSize := len(hash)
	numLimbs := (hashByteSize + uint32ByteSize - 1) / uint32ByteSize // rounds up the division
	for len(hash) != numLimbs*uint32ByteSize {
		hash = append(hash, 0)
	}
	flattenTxTreeRoot := make([]*ffg.Element, numLimbs)
	for i := 0; i < len(flattenTxTreeRoot); i++ {
		v := binary.BigEndian.Uint32(hash[uint32ByteSize*i : uint32ByteSize*(i+1)])
		flattenTxTreeRoot[i] = new(ffg.Element).SetUint64(uint64(v))
	}

	return flattenTxTreeRoot, nil
}

func hashToWeight(myPublicKey *fp.Element, hash []byte) *big.Int {
	p := myPublicKey.Bytes()
	flatten := []byte{}
	flatten = append(flatten, p[:]...)
	flatten = append(flatten, hash...)

	flatten2 := make([]*ffg.Element, len(flatten))
	for i, v := range flatten {
		flatten2[i] = new(ffg.Element).SetUint64(uint64(uint32(v)))
	}
	challenger := goldenposeidon.NewChallenger()
	challenger.ObserveElements(flatten2)

	const (
		bn254OrderByteSize = 32
		uint32ByteSize     = 4
	)
	output := challenger.GetNChallenges(bn254OrderByteSize / uint32ByteSize)

	a := goldenposeidon.FieldElementSliceToBigInt(output)
	return a.BigInt(new(big.Int))
}
