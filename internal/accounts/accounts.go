package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
	isUtils "github.com/prodadidb/go-validation/is/utils"
)

type PublicKey struct {
	Pk *bn254.G1Affine
}

func NewPublicKey(pk *bn254.G1Affine) (*PublicKey, error) {
	publicKey := PublicKey{Pk: pk}
	if err := checkValidPublicKey(&publicKey); err != nil {
		return nil, err
	}
	return &publicKey, nil
}

func (pk *PublicKey) Set(other *PublicKey) *PublicKey {
	pk.Pk = new(bn254.G1Affine).Set(other.Pk)

	return pk
}

func (pk *PublicKey) String() string {
	return hex.EncodeToString(pk.Pk.Marshal())
}

// NewDummyPublicKey returns the point which the x coordinate is 1.
//
// NOTE: If the x coordinate is 0, there is no corresponding y value.
func NewDummyPublicKey() *PublicKey {
	const dummyPublicKeyY = 2
	point := new(bn254.G1Affine)
	point.X.SetOne()
	point.Y.SetInt64(dummyPublicKeyY)

	return &PublicKey{Pk: point}
}

// Add two public keys as elliptic curve points.
func (pk *PublicKey) Add(a, b *PublicKey) *PublicKey {
	if pk.Pk == nil {
		pk.Pk = new(bn254.G1Affine)
	}

	pk.Pk = new(bn254.G1Affine).Add(a.Pk, b.Pk)
	return pk
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

// HexToPrivateKey creates a new PrivateKey instance with a validated private key.
// If the resulting public key is invalid, it returns an error.
func HexToPrivateKey(hexPrivateKey string) (*PrivateKey, error) {
	if !isUtils.IsHexadecimal(hexPrivateKey) {
		return nil, ErrHEXPrivateKeyInvalid
	}

	decKey, err := hex.DecodeString(hexPrivateKey)
	if err != nil {
		return nil, errors.Join(ErrDecodePrivateKeyFail, err)
	}

	var pk *PrivateKey
	pk, err = NewPrivateKey(new(big.Int).SetBytes(decKey))
	if err != nil {
		return nil, errors.Join(ErrCreatePrivateKeyFail, err)
	}

	return pk, nil
}

// NewPrivateKey creates a new PrivateKey instance with a validated private key.
// If the resulting public key is invalid, it returns an error.
func NewPrivateKey(privateKey *big.Int) (*PrivateKey, error) {
	const int0Key = 0
	if privateKey == nil {
		return nil, ErrInputPrivateKeyEmpty
	}
	if privateKey.Cmp(new(big.Int).Mod(privateKey, fr.Modulus())) != int0Key {
		return nil, ErrInputPrivateKeyInvalid
	}
	if privateKey.Cmp(big.NewInt(int0Key)) == int0Key {
		return nil, ErrInputPrivateKeyIsZero
	}

	publicKey := privateKeyToPublicKey(new(big.Int).Set(privateKey))
	if err := checkValidPublicKey(&publicKey); err != nil {
		return nil, errors.Join(ErrValidPublicKeyFail, err)
	}

	a := new(PrivateKey)
	a.sk = privateKey
	a.PublicKey = publicKey

	return a, nil
}

// NewPrivateKeyWithReCalcPubKeyIfPkNegates creates a new PrivateKey instance.
// If the resulting public key is invalid, it negates the private key and recalculates the public key.
// Therefore, the output private key may differ from the input value.
func NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey *big.Int) (*PrivateKey, error) {
	const (
		int0Key = 0
	)

	if privateKey == nil {
		return nil, ErrInputPrivateKeyEmpty
	}
	if new(big.Int).Mod(privateKey, fr.Modulus()).Cmp(big.NewInt(int0Key)) == int0Key {
		return nil, ErrInputPrivateKeyIsZero
	}

	publicKey := privateKeyToPublicKey(privateKey)
	if err := checkValidPublicKey(&publicKey); err != nil {
		privateKey = new(big.Int).Neg(privateKey)
		publicKey = privateKeyToPublicKey(privateKey)
		if err = checkValidPublicKey(&publicKey); err != nil {
			return nil, errors.Join(ErrValidPublicKeyFail, err)
		}
	}

	a := new(PrivateKey)
	a.sk = new(big.Int).Mod(privateKey, fr.Modulus())
	a.PublicKey = publicKey

	return a, nil
}

// checkValidPublicKey verifies that the y coordinate of the given public key is an even number.
// It returns an error if the y coordinate is not even.
func checkValidPublicKey(publicKey *PublicKey) error {
	const (
		int0Key       = 0
		parityDivisor = 2
	)
	publicKeyInt := publicKey.Pk.Y.BigInt(new(big.Int))
	parity := new(big.Int).Mod(publicKeyInt, big.NewInt(parityDivisor))

	// Check if the parity is zero
	if parity.Cmp(big.NewInt(int0Key)) != int0Key {
		return ErrPrivateKeyWithPublicKeyInvalid
	}

	return nil
}

// NewINTMAXAccountFromECDSAKey creates a new PrivateKey instance for an INTMAX account
// from an existing ECDSA private key. It returns an error if the derived private key is invalid,
// but the probability of such an event is extremely low.
func NewINTMAXAccountFromECDSAKey(pk *ecdsa.PrivateKey) (*PrivateKey, error) {
	data := pk.D.Bytes()
	salt := []byte("INTMAX")
	hasher := sha512.New()
	for {
		_, _ = hasher.Write(salt)
		_, _ = hasher.Write(data)
		digest := hasher.Sum(nil)
		privateKey := new(big.Int).SetBytes(digest)
		account, err := NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
		if err != nil {
			continue
		}
		accountAddress := account.ToAddress()
		address := accountAddress.String()
		_, err = NewPublicKeyFromAddressHex(address)
		if err != nil {
			continue
		}
		message := make([]ffg.Element, 1)
		for i := 0; i < len(message); i++ {
			message[i].SetBytes(data)
		}
		sign, err := account.Sign(message)
		if err != nil {
			continue
		}
		var publicKey *PublicKey
		publicKey, err = NewPublicKeyFromAddressHex(address)
		if err != nil {
			continue
		}
		err = VerifySignature(sign, publicKey, message)
		if err != nil {
			continue
		}
		return account, nil
	}
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

// String returns the private key as a hex string.
// It returns a 32-byte hex string without 0x.
func (a *PrivateKey) String() string {
	return hex.EncodeToString(a.Marshal())
}

func NewPrivateKeyFromString(s string) (*PrivateKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	pk := new(PrivateKey)
	err = pk.Unmarshal(b)
	if err != nil {
		return nil, err
	}

	return pk, nil
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

type Address [32]byte

// ToAddress converts the public key to an address.
// It returns a 32-byte hex string with 0x.
func (pk *PublicKey) ToAddress() Address {
	const int32Key = 32
	result := [int32Key]byte{}
	copy(result[:], pk.Pk.X.Marshal())

	return Address(result)
}

func NewPublicKeyFromAddressHex(address string) (*PublicKey, error) {
	recoverAddress, err := NewAddressFromHex(address)
	if err != nil {
		return nil, err
	}

	publicKey, err := recoverAddress.Public()
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func NewPublicKeyFromAddressInt(a *big.Int) (*PublicKey, error) {
	address, err := NewAddressFromBytes(a.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to create new address from bytes: %w", err)
	}

	return address.Public()
}

func (pk *PublicKey) BigInt() *big.Int {
	return pk.Pk.X.BigInt(new(big.Int))
}

func (a *Address) BigInt() (*big.Int, error) {
	pk, err := a.Public()
	if err != nil {
		return nil, err
	}

	return pk.BigInt(), nil
}

// ToAddress converts the private key to an address.
// It returns a 32-byte hex string with 0x.
func (a *PrivateKey) ToAddress() Address {
	return a.Public().ToAddress()
}

func NewAddressFromHex(s string) (Address, error) {
	const int66Key = 66
	if len(s) != int66Key || s[:2] != "0x" {
		return Address{}, ErrAddressInvalid
	}
	b, err := hexutil.Decode(s)
	if err != nil {
		return Address{}, errors.Join(ErrDecodeAddressFail, err)
	}

	return NewAddressFromBytes(b)
}

func NewAddressFromBytes(b []byte) (Address, error) {
	const addressByteSize = 32
	if len(b) != addressByteSize {
		return Address{}, ErrAddressInvalid
	}
	var address Address
	copy(address[:], b)
	return address, nil
}

func (a Address) Public() (*PublicKey, error) {
	const mCompressedSmallest byte = 0b10 << 6
	b := a.Bytes()
	b[0] |= mCompressedSmallest
	point := new(bn254.G1Affine)
	_, err := point.SetBytes(b)
	if err != nil {
		return nil, err
	}

	return NewPublicKey(point)
}

func (a Address) Bytes() []byte {
	return a[:]
}

func (a Address) String() string {
	return hexutil.Encode(a[:])
}

func (a Address) hex() []byte {
	const (
		h0xKey  = "0x"
		int2Key = 2
	)
	var buf [len(a)*int2Key + int2Key]byte
	copy(buf[:int2Key], h0xKey)
	hex.Encode(buf[int2Key:], a[:])
	return buf[:]
}

// Format implements fmt.Formatter.
// Address supports the %v, %s, %q, %x, %X and %d format verbs.
func (a Address) Format(s fmt.State, c rune) {
	switch c {
	case 'v', 's':
		_, _ = s.Write(a.hex())
	case 'q':
		q := []byte{'"'}
		_, _ = s.Write(q)
		_, _ = s.Write(a.hex())
		_, _ = s.Write(q)
	case 'x', 'X':
		enc := a.hex()
		if !s.Flag('#') {
			enc = enc[2:]
		}
		if c == 'X' {
			enc = bytes.ToUpper(enc)
		}
		_, _ = s.Write(enc)
	case 'd':
		_, _ = fmt.Fprint(s, [len(a)]byte(a))
	default:
		_, _ = fmt.Fprintf(s, "%%!%c(address=%x)", c, a)
	}
}
