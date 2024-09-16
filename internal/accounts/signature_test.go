package accounts_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/mnemonic_wallet"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureByINTMAXAccount(t *testing.T) {
	t.Parallel()

	// Generate key pairs for both parties
	const (
		mnPassword = ""
		derivation = "m/44'/60'/0'/0/0"
	)

	w, err := mnemonic_wallet.New().WalletGenerator(derivation, mnPassword)
	assert.NoError(t, err)

	pk, err := intMaxAcc.HexToPrivateKey(w.IntMaxPrivateKey)
	assert.NoError(t, err)

	keyPair, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(pk.BigInt())
	assert.NoError(t, err)

	messageHex := "99947e33d5d672d82b7f221f4899e31b574314692b3ecd6a01693a7c38af1271"
	messageBytes, err := hex.DecodeString(messageHex)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(messageBytes))

	flattenMessage := finite_field.BytesToFieldElementSlice(messageBytes)

	signature, err := keyPair.Sign(flattenMessage)
	assert.NoError(t, err)

	addr := keyPair.Public().ToAddress().String()
	pubKey, err := intMaxAcc.NewPublicKeyFromAddressHex(addr)
	assert.NoError(t, err)

	err = intMaxAcc.VerifySignature(signature, keyPair.Public(), flattenMessage)
	assert.NoError(t, err)

	err = intMaxAcc.VerifySignature(signature, pubKey, flattenMessage)
	assert.NoError(t, err)
}

func TestVerifyHexSignature(t *testing.T) {
	addressHex := "0x0ec939a62c909fd83c2af088f24482ae9057f4450a22f6b7d9d3038536356d95"
	address, err := intMaxAcc.NewAddressFromHex(addressHex)
	assert.NoError(t, err)
	publicKey, err := address.Public()
	assert.NoError(t, err)

	signature, err := intMaxAcc.DecodeG2CurvePoint("0367810768faa199d6cb0d4ab3f5ccb0b40ee77197512fb06225b62455e94e0a0c8927695b7e3644306fa82de066df3472784dc7cf9c19ca172141bdf74d4cb9089255ec6a86312ac1d6f6c43c90abc82547f3a47bd8bc6ab68574037c8f271c015c7faefff4762e6d25b17060dd6b8a4ac44a3190fd7096514d4169c98d15e0")
	assert.NoError(t, err)

	messageHex := "99947e33d5d672d82b7f221f4899e31b574314692b3ecd6a01693a7c38af1271"
	messageBytes, err := hex.DecodeString(messageHex)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(messageBytes))

	flattenMessage := finite_field.BytesToFieldElementSlice(messageBytes)
	err = intMaxAcc.VerifySignature(signature, publicKey, flattenMessage)
	assert.NoError(t, err)
}

func TestAggregatedPublicKey(t *testing.T) {
	privateKey, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	assert.NoError(t, err)

	publicKey := privateKey.Public()
	fmt.Printf("publicKey: %v\n", publicKey.ToAddress().String())
	publicKeysHash, err := hexutil.Decode("0xad5d1d11ba412b3ea4aed201704872794628d2ce09d4bb3e0777cced104f389e")
	require.NoError(t, err)
	weightedPublicKey := publicKey.WeightByHash(publicKeysHash)
	fmt.Printf("weightedPublicKey: %v\n", weightedPublicKey.BigInt())
}

func TestAggregatedSignature(t *testing.T) {
	t.Parallel()

	// Generate key pairs for both parties.
	keyPairs := make([]*intMaxAcc.PrivateKey, 3)
	for i := 0; i < len(keyPairs); i++ {
		privateKey, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
		assert.NoError(t, err)
		privateKey.Add(privateKey, big.NewInt(1))
		keyPair, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
		assert.NoError(t, err)
		keyPairs[i] = keyPair
	}

	txTreeRoot := [32]byte{}
	rand.Read(txTreeRoot[:])
	flattenTxTreeRoot := finite_field.BytesToFieldElementSlice(txTreeRoot[:])

	publicKeysHash := []byte("publicKeysHash") // dummy
	weightedkeyPairs := make([]*intMaxAcc.PrivateKey, len(keyPairs))
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

	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for _, keyPair := range keyPairs {
		weightedPublicKey := keyPair.Public().WeightByHash(publicKeysHash)
		aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
	}

	err := intMaxAcc.VerifySignature(aggregatedSignature, aggregatedPublicKey, flattenTxTreeRoot)
	assert.NoError(t, err)
}
