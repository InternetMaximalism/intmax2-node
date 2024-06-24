package types

import (
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"sort"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyBlockValidation(t *testing.T) {
	keyPairs := make([]*accounts.PrivateKey, 128)
	for i := 0; i < len(keyPairs); i++ {
		keyPairs[i] = accounts.GenerateKey()
	}

	// Sort by x-coordinate of public key
	sort.Slice(keyPairs, func(i, j int) bool {
		return keyPairs[i].Pk.X.Cmp(&keyPairs[j].Pk.X) > 0
	})

	senders := make([]Sender, 128)
	for i, keyPair := range keyPairs {
		senders[i] = Sender{
			PublicKey: keyPair.Public(),
			AccountID: 0,
			IsSigned:  true,
		}
	}
	// default is the point which x is 1
	defaultPublicKey := accounts.NewPublicKey(new(bn254.G1Affine))
	defaultPublicKey.Pk.X.SetOne()
	defaultPublicKey.Pk.Y.SetZero() // NOTE: This is not a valid public key
	for i := len(keyPairs); i < len(senders); i++ {
		senders[i] = Sender{
			PublicKey: defaultPublicKey,
			AccountID: 0,
			IsSigned:  false,
		}
	}

	txRoot, err := new(poseidonHashOut).SetRandom()
	assert.NoError(t, err)

	senderPublicKeys := make([]byte, len(senders)*numPublicKeyBytes)
	for i, pk := range senders {
		senderPublicKey := pk.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeys)
	aggregatedPublicKey := accounts.NewPublicKey(new(bn254.G1Affine))
	for _, sender := range senders {
		aggregatedPublicKey.Pk.Add(aggregatedPublicKey.Pk, sender.PublicKey.WeightByHash(publicKeysHash).Pk)
	}

	message := finite_field.BytesToFieldElementSlice(txRoot.Marshal())

	aggregatedSignature := new(bn254.G2Affine)
	for _, keyPair := range keyPairs {
		signature, err := keyPair.WeightByHash(publicKeysHash).Sign(message)
		assert.NoError(t, err)
		aggregatedSignature.Add(aggregatedSignature, signature)
	}

	blockContent := NewBlockContent(
		PublicKeySenderType,
		senders,
		*txRoot,
		aggregatedSignature,
	)
	assert.NoError(t, blockContent.IsValid())
}

func TestAccountIDBlockValidation(t *testing.T) {
	keyPairs := make([]*accounts.PrivateKey, 1)
	for i := 0; i < len(keyPairs); i++ {
		keyPairs[i] = accounts.GenerateKey()
	}

	// Sort by x-coordinate of public key
	sort.Slice(keyPairs, func(i, j int) bool {
		return keyPairs[i].Pk.X.Cmp(&keyPairs[j].Pk.X) > 0
	})

	senders := make([]Sender, 2)
	for i, keyPair := range keyPairs {
		senders[i] = Sender{
			PublicKey: keyPair.Public(),
			AccountID: uint64(i) + 1,
			IsSigned:  true,
		}
	}
	// default is the point which x is 1
	defaultPublicKey := accounts.NewPublicKey(new(bn254.G1Affine))
	defaultPublicKey.Pk.X.SetOne()
	defaultPublicKey.Pk.Y.SetZero() // NOTE: This is not a valid public key
	for i := len(keyPairs); i < len(senders); i++ {
		senders[i] = Sender{
			PublicKey: defaultPublicKey,
			AccountID: 0,
			IsSigned:  false,
		}
	}

	txRoot, err := new(poseidonHashOut).SetRandom()
	assert.NoError(t, err)

	senderPublicKeys := make([]byte, len(senders)*numPublicKeyBytes)
	for i, sender := range senders {
		if sender.IsSigned {
			senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
		}
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeys)
	aggregatedPublicKey := accounts.NewPublicKey(new(bn254.G1Affine))
	for _, sender := range senders {
		if sender.IsSigned {
			aggregatedPublicKey.Pk.Add(aggregatedPublicKey.Pk, sender.PublicKey.WeightByHash(publicKeysHash).Pk)
		}
	}

	message := finite_field.BytesToFieldElementSlice(txRoot.Marshal())

	aggregatedSignature := new(bn254.G2Affine)
	for i, keyPair := range keyPairs {
		if senders[i].IsSigned {
			signature, err := keyPair.WeightByHash(publicKeysHash).Sign(message)
			assert.NoError(t, err)
			aggregatedSignature.Add(aggregatedSignature, signature)
		}
	}

	blockContent := NewBlockContent(
		AccountIDSenderType,
		senders,
		*txRoot,
		aggregatedSignature,
	)
	assert.NoError(t, blockContent.IsValid())
}
