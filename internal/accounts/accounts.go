package accounts

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha512"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type PublicKey struct {
	Pk *bn254.G1Affine
}

func NewPublicKey(pk *bn254.G1Affine) *PublicKey {
	return &PublicKey{Pk: pk}
}

type PrivateKey struct {
	PublicKey
	sk *big.Int
}

func privateKeyToPublicKey(privateKey *big.Int) PublicKey {
	return PublicKey{Pk: new(bn254.G1Affine).ScalarMultiplicationBase(privateKey)}
}

func NewPrivateKey(privateKey *big.Int) *PrivateKey {
	// Calculate the public key.
	// Multiply generator of G1 with private key.
	publicKey := privateKeyToPublicKey(privateKey)

	a := new(PrivateKey)
	a.sk = privateKey
	a.PublicKey = publicKey

	return a
}

func GenerateKey() (*PrivateKey, error) {
	// Generate a random private key.
	// Private key is a random number between 0 and the order of the scalar field.
	privateKey, err := rand.Int(rand.Reader, fr.Modulus())
	if err != nil {
		return nil, err
	}

	return NewPrivateKey(privateKey), nil
}

func NewINTMAXAccountFromECDSAKey(pk *ecdsa.PrivateKey) *PrivateKey {
	data := pk.D.Bytes()
	salt := []byte("INTMAX")

	hasher := sha512.New()
	_, _ = hasher.Write(salt)
	_, _ = hasher.Write(data)
	digest := hasher.Sum(nil)

	// privateKey = digest (mod order)
	privateKey := new(big.Int).SetBytes(digest)

	return NewPrivateKey(privateKey)
}

// ECDH calculates the shared secret between my private key and partner's public key.
func (a *PrivateKey) ECDH(partnerKey *PublicKey) *bn254.G1Affine {
	return new(bn254.G1Affine).ScalarMultiplication(partnerKey.Pk, a.sk)
}

func (a *PrivateKey) Public() *PublicKey {
	return &a.PublicKey
}

func (a *PrivateKey) Equal(other *PrivateKey) bool {
	return a.sk.Cmp(other.sk) == 0
}

func (a *PrivateKey) Marshal() []byte {
	return a.sk.Bytes()
}

func (a *PrivateKey) Unmarshal(buf []byte) error {
	a.sk = new(big.Int).SetBytes(buf)
	a.PublicKey = privateKeyToPublicKey(a.sk)
	return nil
}

func (a *PrivateKey) BigInt() *big.Int {
	return a.sk
}

func (a *PrivateKey) String() string {
	return a.sk.String()
}

func (pk *PublicKey) Equal(other *PublicKey) bool {
	return pk.Pk.Equal(other.Pk)
}

func (pk *PublicKey) Marshal() []byte {
	return pk.Pk.Marshal()
}

func (pk *PublicKey) Unmarshal(buf []byte) error {
	publicKey := new(bn254.G1Affine)
	err := publicKey.Unmarshal(buf)
	if err != nil {
		return err
	}

	pk.Pk = publicKey
	return nil
}

func (pk *PublicKey) G1Affine() *bn254.G1Affine {
	return pk.Pk
}

func (pk *PublicKey) String() string {
	return pk.Pk.String()
}
