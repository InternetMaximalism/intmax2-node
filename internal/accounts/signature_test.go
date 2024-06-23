package accounts

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
)

func TestSignatureByINTMAXAccount(t *testing.T) {
	t.Parallel()

	// Generate key pairs for both parties
	keyPair := GenerateKey()

	message := make([]*ffg.Element, 20)
	for i := 0; i < len(message); i++ {
		message[i] = new(ffg.Element).SetUint64(uint64(i))
	}
	signature, err := keyPair.Sign(message)
	assert.NoError(t, err)

	err = VerifySignature(signature, keyPair.Public(), message)
	assert.NoError(t, err)
}

func TestVerifyHexSignature(t *testing.T) {
	publicKey := new(PublicKey)
	bytes, err := hex.DecodeString("22a0e14b45128c24d8deb2c336de9250ee07357b0b755518e6a8e42058b58a4f10d4f0cea7a7b3996cd919fae2aae91df8171b3bf0cb7c13af68fc1f3038d5f7")
	assert.NoError(t, err)
	err = publicKey.Unmarshal(bytes)
	assert.NoError(t, err)
	signature, err := DecodeG2CurvePoint("27f811fe50964adcb0345ddf85dd0e2e913229991b1d2a551df2908e8ccd3bfc2ba7d3c0ce4096f524d22afeba96b6ce95a6357b5336f9cc57dc0cc78fa605e604781cec49a668fc7ec5dc22fd5f9e49e2b594b1ff9b8067c97d2b60d6be6cd0048da9489637392dc5c427d7b5e9b0976158a3f06b58820c90245ad68675b8b4")
	assert.NoError(t, err)

	messageHex := "99947e33d5d672d82b7f221f4899e31b574314692b3ecd6a01693a7c38af1271"
	messageBytes, err := hex.DecodeString(messageHex)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(messageBytes))

	flattenMessage := make([]*ffg.Element, len(messageBytes)/4)
	for i := 0; i < len(flattenMessage); i++ {
		// big endian
		limb := binary.BigEndian.Uint32(messageBytes[4*i : 4*(i+1)])
		flattenMessage[i] = new(ffg.Element).SetUint64(uint64(limb))
	}

	err = VerifySignature(signature, publicKey, flattenMessage)
	assert.NoError(t, err)

}

func TestAggregatedSignature(t *testing.T) {
	t.Parallel()

	// Generate key pairs for both parties.
	keyPairs := make([]*PrivateKey, 3)
	for i := 0; i < len(keyPairs); i++ {
		keyPair := GenerateKey()
		keyPairs[i] = keyPair
	}

	txTreeRoot := [32]byte{}
	rand.Read(txTreeRoot[:])
	flattenTxTreeRoot, err := HashToFieldElementSlice(txTreeRoot[:])
	assert.NoError(t, err)

	publicKeysHash := []byte("publicKeysHash") // dummy
	weightedkeyPairs := make([]*PrivateKey, len(keyPairs))
	for i, keyPair := range keyPairs {
		weightedkeyPairs[i] = keyPair.WeightByHash(publicKeysHash)
	}

	signatures := make([]*bn254.G2Affine, len(keyPairs))
	for i, keyPair := range keyPairs {
		signature, err := keyPair.WeightByHash(publicKeysHash).Sign(flattenTxTreeRoot)
		assert.NoError(t, err)
		signatures[i] = signature
	}

	aggregatedSignature := new(bn254.G2Affine)
	for _, signature := range signatures {
		aggregatedSignature.Add(aggregatedSignature, signature)
	}

	aggregatedPublicKey := new(bn254.G1Affine)
	for _, keyPair := range keyPairs {
		weightedPublicKey := keyPair.Public().WeightByHash(publicKeysHash)
		aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey.Pk)
	}

	err = VerifySignature(aggregatedSignature, NewPublicKey(aggregatedPublicKey), flattenTxTreeRoot)
	assert.NoError(t, err)
}
