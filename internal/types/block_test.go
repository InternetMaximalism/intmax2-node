package types_test

import (
	"crypto/rand"
	"fmt"
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"sort"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyBlockValidation(t *testing.T) {
	keyPairs := make([]*accounts.PrivateKey, 100)
	for i := 0; i < len(keyPairs); i++ {
		privateKey, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
		require.NoError(t, err)
		privateKey.Add(privateKey, big.NewInt(1))
		keyPairs[i], err = accounts.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
		require.NoError(t, err)
	}

	// Sort by x-coordinate of public key
	sort.Slice(keyPairs, func(i, j int) bool {
		return keyPairs[i].Pk.X.Cmp(&keyPairs[j].Pk.X) > 0
	})

	for i := 1; i < len(keyPairs); i++ {
		require.True(t, keyPairs[i-1].Pk.X.Cmp(&keyPairs[i].Pk.X) > 0)
	}

	senders := make([]intMaxTypes.Sender, 128)
	for i, keyPair := range keyPairs {
		senders[i] = intMaxTypes.Sender{
			PublicKey: keyPair.Public(),
			AccountID: 0,
			IsSigned:  randomBool(),
		}
	}

	defaultSender := intMaxTypes.NewDummySender()
	for i := len(keyPairs); i < len(senders); i++ {
		senders[i] = defaultSender
	}

	txRoot, err := new(intMaxTypes.PoseidonHashOut).SetRandom()
	require.NoError(t, err)

	const numOfSenders = 128
	senderPublicKeysBytes := make([]byte, numOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, pk := range senders {
		senderPublicKey := pk.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[32*i:32*(i+1)], senderPublicKey[:])
	}
	for i := len(senders); i < numOfSenders; i++ {
		senderPublicKey := defaultSender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[32*i:32*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(accounts.PublicKey)
	for _, sender := range senders {
		if sender.IsSigned {
			aggregatedPublicKey.Add(aggregatedPublicKey, sender.PublicKey.WeightByHash(publicKeysHash))
		}
	}

	message := finite_field.BytesToFieldElementSlice(txRoot.Marshal())

	aggregatedSignature := new(bn254.G2Affine)
	for i, keyPair := range keyPairs {
		if senders[i].IsSigned {
			signature, err := keyPair.WeightByHash(publicKeysHash).Sign(message)
			require.NoError(t, err)
			aggregatedSignature.Add(aggregatedSignature, signature)
		}
	}

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.PublicKeySenderType,
		senders,
		*txRoot,
		aggregatedSignature,
	)
	err = blockContent.IsValid()
	require.NoError(t, err)

	_, err = intMaxTypes.MakePostRegistrationBlockInput(
		blockContent,
	)
	require.NoError(t, err)
}

func TestAccountIDBlockValidation(t *testing.T) {
	keyPairs := make([]*accounts.PrivateKey, 100)
	for i := 0; i < len(keyPairs); i++ {
		privateKey, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
		require.NoError(t, err)
		privateKey.Add(privateKey, big.NewInt(1))
		keyPairs[i], err = accounts.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
		require.NoError(t, err)
	}

	// Sort by x-coordinate of public key
	sort.Slice(keyPairs, func(i, j int) bool {
		return keyPairs[i].Pk.X.Cmp(&keyPairs[j].Pk.X) > 0
	})

	for i := 1; i < len(keyPairs); i++ {
		require.True(t, keyPairs[i-1].Pk.X.Cmp(&keyPairs[i].Pk.X) > 0)
	}

	senders := make([]intMaxTypes.Sender, 128)
	for i, keyPair := range keyPairs {
		senders[i] = intMaxTypes.Sender{
			PublicKey: keyPair.Public(),
			AccountID: uint64(i) + 2,
			IsSigned:  randomBool(),
		}
	}

	defaultSender := intMaxTypes.NewDummySender()
	for i := len(keyPairs); i < len(senders); i++ {
		senders[i] = defaultSender
	}

	txRoot, err := new(intMaxTypes.PoseidonHashOut).SetRandom()
	require.NoError(t, err)

	const numOfSenders = 128
	senderPublicKeys := make([]byte, numOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senders {
		senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
	}
	for i := len(senders); i < numOfSenders; i++ {
		senderPublicKey := defaultSender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeys)
	aggregatedPublicKey := new(accounts.PublicKey)
	for _, sender := range senders {
		if sender.IsSigned {
			aggregatedPublicKey.Add(aggregatedPublicKey, sender.PublicKey.WeightByHash(publicKeysHash))
		}
	}

	message := finite_field.BytesToFieldElementSlice(txRoot.Marshal())

	aggregatedSignature := new(bn254.G2Affine)
	for i, keyPair := range keyPairs {
		if senders[i].IsSigned {
			signature, err := keyPair.WeightByHash(publicKeysHash).Sign(message)
			require.NoError(t, err)
			aggregatedSignature.Add(aggregatedSignature, signature)
		}
	}

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.AccountIDSenderType,
		senders,
		*txRoot,
		aggregatedSignature,
	)
	require.NoError(t, blockContent.IsValid())

	_, err = intMaxTypes.MakePostRegistrationBlockInput(
		blockContent,
	)
	require.NoError(t, err)
}

func TestMarshalAccountIds(t *testing.T) {
	accountIds := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	encodedAccountIds, err := intMaxTypes.MarshalAccountIds(accountIds)
	require.NoError(t, err)

	fmt.Printf("Marshaled account IDs: %x\n", encodedAccountIds)

	decodedAccountIds, err := intMaxTypes.UnmarshalAccountIds(encodedAccountIds)
	require.NoError(t, err)

	require.Equal(t, accountIds, decodedAccountIds)
}

func randomBool() bool {
	var b [1]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}

	return b[0]%2 == 0
}
