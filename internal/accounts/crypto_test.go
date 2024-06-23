package accounts

import (
	"crypto/aes"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestECDHKeyExchangeAndAES(t *testing.T) {
	t.Parallel()

	// Generate key pairs for both parties
	keyPairA := GenerateKey()
	keyPairB := GenerateKey()

	// Generate shared secret
	sharedSecretA := keyPairA.ECDH(keyPairB.Public())

	// Should be the same as above
	sharedSecretB := keyPairB.ECDH(keyPairA.Public())
	assert.Equal(t, sharedSecretA, sharedSecretB, "wrong shared secrets")

	aesKey := sha256.Sum256(sharedSecretA.Marshal())

	plaintext := []byte("This is a secret message.")
	iv, ciphertext, err := EncryptAES(aesKey[:], plaintext)
	assert.NoError(t, err)

	decrypted, err := DecryptAES(aesKey[:], iv, ciphertext[aes.BlockSize:])
	assert.NoError(t, err)
	assert.Equal(t, string(decrypted), string(plaintext))
}
