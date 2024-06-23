package accounts

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha512"
	"errors"
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

// Calculate the public key.
// Multiply generator of G1 with private key.
func privateKeyToPublicKey(privateKey *big.Int) PublicKey {
	return PublicKey{Pk: new(bn254.G1Affine).ScalarMultiplicationBase(privateKey)}
}

// NewPrivateKey creates a new PrivateKey instance with a validated private key.
// If the resulting public key is invalid, it returns an error.
func NewPrivateKey(privateKey *big.Int) (*PrivateKey, error) {
	if privateKey == nil {
		return nil, errors.New("private key should not be nil")
	}
	if privateKey.Cmp(new(big.Int).Mod(privateKey, fr.Modulus())) != 0 {
		return nil, errors.New("private key should be less than the order of the scalar field")
	}
	if privateKey.Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New("private key should not be zero")
	}

	publicKey := privateKeyToPublicKey(privateKey)
	if err := checkValidPublicKey(&publicKey); err != nil {
		return nil, err
	}

	a := new(PrivateKey)
	a.sk = privateKey
	a.PublicKey = publicKey

	return a, nil
}

// newPrivateKey creates a new PrivateKey instance.
// If the resulting public key is invalid, it negates the private key and recalculates the public key.
// Therefore, the output private key may differ from the input value.
func newPrivateKey(privateKey *big.Int) (*PrivateKey, error) {
	if privateKey == nil {
		return nil, errors.New("private key should not be nil")
	}
	if new(big.Int).Mod(privateKey, fr.Modulus()).Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New("private key should not be zero")
	}

	publicKey := privateKeyToPublicKey(privateKey)
	if err := checkValidPublicKey(&publicKey); err != nil {
		privateKey = new(big.Int).Neg(privateKey)
		publicKey = privateKeyToPublicKey(privateKey)
		if err = checkValidPublicKey(&publicKey); err != nil {
			return nil, errors.New("invalid private key: the y coordinate of public key should be even number")
		}
	}

	a := new(PrivateKey)
	a.sk = new(big.Int).Mod(privateKey, fr.Modulus())
	a.PublicKey = publicKey

	return a, nil
}

func checkValidPublicKey(publicKey *PublicKey) error {
	// The y coordinate of the public key should be even number.
	publicKeyInt := publicKey.Pk.Y.BigInt(new(big.Int))
	const parityDivisor = 2
	parity := new(big.Int).Mod(publicKeyInt, big.NewInt(parityDivisor))
	if parity.Cmp(big.NewInt(0)) != 0 {
		return errors.New("invalid private key: the y coordinate of public key should be even number")
	}

	return nil
}

func GenerateKey() *PrivateKey {
	// Generate a random private key.
	// Private key is a random number between 1 and the order of the scalar field.
	privateKey, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
	if err != nil {
		panic(err)
	}

	privateKey.Add(privateKey, big.NewInt(1))

	a, err := newPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	return a
}

func NewINTMAXAccountFromECDSAKey(pk *ecdsa.PrivateKey) (*PrivateKey, error) {
	data := pk.D.Bytes()
	salt := []byte("INTMAX")

	hasher := sha512.New()
	_, _ = hasher.Write(salt)
	_, _ = hasher.Write(data)
	digest := hasher.Sum(nil)

	// privateKey = digest (mod order)
	privateKey := new(big.Int).SetBytes(digest)

	return newPrivateKey(privateKey)
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

func (pk *PublicKey) String() string {
	return pk.Pk.String()
}
