package accounts

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

func TestShouldNotGenerateNilAccount(t *testing.T) {
	t.Parallel()

	a, err := NewPrivateKey(nil)
	expectedError := "private key should not be nil"
	assert.Equal(t, expectedError, err.Error())
	assert.Nil(t, a)
}

func TestShouldNotGenerateZeroAccount(t *testing.T) {
	t.Parallel()

	a, err := NewPrivateKey(big.NewInt(0))
	expectedError := "private key should not be zero"
	assert.Equal(t, expectedError, err.Error())
	assert.Nil(t, a)
}

func TestShouldNotGenerateInvalidAccount(t *testing.T) {
	t.Parallel()

	a, err := NewPrivateKey(big.NewInt(3))
	expectedError := "invalid private key: the y coordinate of public key should be even number"
	assert.Equal(t, expectedError, err.Error())
	assert.Nil(t, a)
}

func TestNewPrivateKey(t *testing.T) {
	t.Parallel()

	_, err := NewPrivateKey(big.NewInt(2))
	assert.NoError(t, err)
}

func TestRegenerateAccount(t *testing.T) {
	t.Parallel()

	p := big.NewInt(-2)
	a, err := newPrivateKey(p)

	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(2), a.BigInt())
}

func TestNewINTMAXAccountFromEthereumKey(t *testing.T) {
	t.Parallel()

	ethereumPrivateKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	assert.NoError(t, err)

	_, err = NewINTMAXAccountFromECDSAKey(ethereumPrivateKey)
	assert.NoError(t, err)
}

func TestMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	account := GenerateKey()
	marshaled := account.Public().Marshal()

	var publicKey PublicKey
	publicKey.Unmarshal(marshaled)

	assert.True(t, publicKey.Equal(account.Public()))
}
