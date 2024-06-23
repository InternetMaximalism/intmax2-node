package accounts

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func TestNewINTMAXAccountFromEthereumKey(t *testing.T) {
	t.Parallel()

	ethereumPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Error generating key: %v", err)
	}

	account := NewINTMAXAccountFromECDSAKey(ethereumPrivateKey)

	fmt.Printf("public key: %x\n", account.Public().Marshal())
}

func TestMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	ethereumPrivateKey, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		t.Fatalf("Error generating key: %v", err)
	}

	account := NewINTMAXAccountFromECDSAKey(ethereumPrivateKey)
	marshaled := account.Public().Marshal()

	var publicKey PublicKey
	publicKey.Unmarshal(marshaled)

	if !publicKey.Equal(account.Public()) {
		t.Fatalf("public keys are not equal")
	}
}
