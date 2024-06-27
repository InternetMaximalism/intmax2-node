package accounts_test

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	intMaxAcc "intmax2-node/internal/accounts"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/stretchr/testify/assert"
)

func TestECDHKeyExchangeAndAES(t *testing.T) {
	t.Parallel()

	// Generate key pairs for both parties
	privateKeyA, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
	assert.NoError(t, err)
	privateKeyA.Add(privateKeyA, big.NewInt(1))
	keyPairA, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKeyA)
	assert.NoError(t, err)
	privateKeyB, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
	assert.NoError(t, err)
	privateKeyB.Add(privateKeyB, big.NewInt(1))
	keyPairB, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKeyB)
	assert.NoError(t, err)

	// Generate shared secret
	sharedSecretA := keyPairA.ECDH(keyPairB.Public())

	// Should be the same as above
	sharedSecretB := keyPairB.ECDH(keyPairA.Public())
	assert.Equal(t, sharedSecretA, sharedSecretB, "wrong shared secrets")

	aesKey := sha256.Sum256(sharedSecretA.Marshal())

	plaintext := []byte("This is a secret message.")
	iv, ciphertext, err := intMaxAcc.EncryptAES(aesKey[:], plaintext)
	assert.NoError(t, err)

	decrypted, err := intMaxAcc.DecryptAES(aesKey[:], iv, ciphertext[aes.BlockSize:])
	assert.NoError(t, err)
	assert.Equal(t, string(decrypted), string(plaintext))
}
