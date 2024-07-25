package accounts_test

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	intMaxAcc "intmax2-node/internal/accounts"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/crypto/ecies"
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
	iv, ciphertext, err := intMaxAcc.EncryptWithAES(aesKey[:], plaintext)
	assert.NoError(t, err)

	decrypted, err := intMaxAcc.DecryptWithAES(aesKey[:], iv, ciphertext[aes.BlockSize:])
	assert.NoError(t, err)
	assert.Equal(t, string(decrypted), string(plaintext))
}

func TestAEADEncryption(t *testing.T) {
	key := make([]byte, 32) // 32 bytes for AES-256
	_, err := rand.Read(key)
	assert.NoError(t, err)

	plaintext := []byte("Hello, World!")
	nonce, ciphertext, err := intMaxAcc.EncryptWithAEAD(key, plaintext)
	assert.NoError(t, err)

	decryptedText, err := intMaxAcc.DecryptWithAEAD(key, nonce, ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, string(plaintext), string(decryptedText))
}

func TestECIES(t *testing.T) {
	// definition of the BN254 curve
	bn254G1Curve := &intMaxAcc.BN254G1Curve{}

	prv1, err := ecies.GenerateKey(rand.Reader, bn254G1Curve, ecies.ECIES_AES128_SHA256)
	assert.NoError(t, err)

	wallet2, err := intMaxAcc.NewPrivateKey(big.NewInt(5))
	assert.NoError(t, err)

	message := []byte("Hello, world.")
	ct, err := intMaxAcc.EncryptECIES(rand.Reader, wallet2.Public(), message)
	assert.NoError(t, err)

	const (
		publicKeyLen  = 65 // size of uncompressed public key
		messageTagLen = 32 // output size of SHA-256
	)
	ciphertextLen := ecies.ECIES_AES128_SHA256.BlockSize + len(message)
	ctLen := publicKeyLen + ciphertextLen + messageTagLen
	assert.Equal(t, ctLen, len(ct), "ciphertext length is wrong")

	pt, err := wallet2.DecryptECIES(ct)
	assert.NoError(t, err)
	assert.Equal(t, pt, message, "plaintext doesn't match message")

	_, err = prv1.Decrypt(ct, nil, nil)
	assert.Error(t, err, "encryption should not have succeeded")
}
