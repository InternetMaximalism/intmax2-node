package accounts_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

func TestShouldNotGenerateNilAccount(t *testing.T) {
	t.Parallel()

	a, err := intMaxAcc.NewPrivateKey(nil)
	assert.True(t, errors.Is(err, intMaxAcc.ErrInputPrivateKeyEmpty))
	assert.Nil(t, a)
}

func TestShouldNotGenerateZeroAccount(t *testing.T) {
	t.Parallel()

	a, err := intMaxAcc.NewPrivateKey(big.NewInt(0))
	assert.True(t, errors.Is(err, intMaxAcc.ErrInputPrivateKeyIsZero))
	assert.Nil(t, a)
}

func TestShouldNotGenerateInvalidAccount(t *testing.T) {
	t.Parallel()

	a, err := intMaxAcc.NewPrivateKey(big.NewInt(3))
	assert.True(t, errors.Is(err, intMaxAcc.ErrPrivateKeyWithPublicKeyInvalid))
	assert.Nil(t, a)
}

func TestHexToPrivateKey(t *testing.T) {
	t.Parallel()

	pk, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	assert.NoError(t, err)
	assert.NotNil(t, pk)

	var pk2 *intMaxAcc.PrivateKey
	pk2, err = intMaxAcc.HexToPrivateKey(hexutil.Encode(pk.BigInt().Bytes())[2:])
	assert.NoError(t, err)
	assert.NotNil(t, pk2)

	assert.Equal(t, 0, bytes.Compare(pk.PublicKey.Pk.Marshal(), pk2.PublicKey.Pk.Marshal()))
	assert.Equal(t, 0, bytes.Compare(pk.Pk.Marshal(), pk2.Pk.Marshal()))
}

func TestNewPrivateKey(t *testing.T) {
	t.Parallel()

	_, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	assert.NoError(t, err)
}

func TestRegenerateAccount(t *testing.T) {
	t.Parallel()

	p := big.NewInt(-2)
	a, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(p)

	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(2), a.BigInt())
}

func TestNewINTMAXAccountFromEthereumKey(t *testing.T) {
	t.Parallel()

	ethereumPrivateKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	assert.NoError(t, err)

	_, err = intMaxAcc.NewINTMAXAccountFromECDSAKey(ethereumPrivateKey)
	assert.NoError(t, err)
}

func TestMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	privateKey, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
	assert.NoError(t, err)
	privateKey.Add(privateKey, big.NewInt(1))
	account, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
	assert.NoError(t, err)
	marshaled := account.Public().Marshal()

	var publicKey intMaxAcc.PublicKey
	err = publicKey.Unmarshal(marshaled)
	assert.NoError(t, err)

	assert.True(t, publicKey.Equal(account.Public()))
}
