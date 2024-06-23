package accounts

import (
	"crypto/aes"
	"crypto/sha256"
	"testing"
)

func TestECDHKeyExchangeAndAES(t *testing.T) {
	t.Parallel()

	// Generate key pairs for both parties
	keyPairA, err := GenerateKey()
	if err != nil {
		t.Fatalf("Error generating keys: %v", err)
	}

	keyPairB, err := GenerateKey()
	if err != nil {
		t.Fatalf("Error generating keys: %v", err)
	}

	sharedSecretA := keyPairA.ECDH(keyPairB.Public())

	sharedSecretB := keyPairB.ECDH(keyPairA.Public())
	if !sharedSecretA.Equal(sharedSecretB) {
		t.Fatalf("wrong shared secrets")
	}

	aesKey := sha256.Sum256(sharedSecretA.Marshal())

	plaintext := []byte("This is a secret message.")
	iv, ciphertext, err := EncryptAES(aesKey[:], plaintext)
	if err != nil {
		t.Fatalf("Error encrypting message: %v", err)
	}

	decrypted, err := DecryptAES(aesKey[:], iv, ciphertext[aes.BlockSize:])
	if err != nil {
		t.Fatalf("Error decrypting message: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("wrong decrypted message")
	}
}
