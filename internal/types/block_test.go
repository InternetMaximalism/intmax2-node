package types_test

import (
	"crypto/rand"
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"sort"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyBlockValidation(t *testing.T) {
	keyPairs := make([]*accounts.PrivateKey, 10)
	for i := 0; i < len(keyPairs); i++ {
		privateKey, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
		assert.NoError(t, err)
		privateKey.Add(privateKey, big.NewInt(1))
		keyPairs[i], err = accounts.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
		assert.NoError(t, err)
	}

	// Sort by x-coordinate of public key
	sort.Slice(keyPairs, func(i, j int) bool {
		return keyPairs[i].Pk.X.Cmp(&keyPairs[j].Pk.X) > 0
	})

	senders := make([]intMaxTypes.Sender, 128)
	for i, keyPair := range keyPairs {
		senders[i] = intMaxTypes.Sender{
			PublicKey: keyPair.Public(),
			AccountID: 0,
			IsSigned:  true,
		}
	}

	defaultPublicKey := accounts.NewDummyPublicKey()
	for i := len(keyPairs); i < len(senders); i++ {
		senders[i] = intMaxTypes.Sender{
			PublicKey: defaultPublicKey,
			AccountID: 0,
			IsSigned:  false,
		}
	}

	txRoot, err := new(intMaxTypes.PoseidonHashOut).SetRandom()
	assert.NoError(t, err)

	senderPublicKeys := make([]byte, len(senders)*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senders {
		if sender.IsSigned {
			senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
		}
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
			assert.NoError(t, err)
			aggregatedSignature.Add(aggregatedSignature, signature)
		}
	}

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.PublicKeySenderType,
		senders,
		*txRoot,
		aggregatedSignature,
	)
	assert.NoError(t, blockContent.IsValid())

	_, err = intMaxTypes.MakePostRegistrationBlockInput(
		blockContent,
	)
	assert.NoError(t, err)

	// Usage of PostRegistrationBlock:
	// tx, err := intMaxTypes.PostRegistrationBlock(cfg, blockContent)
	// assert.NoError(t, err)

	// fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
}

func TestAccountIDBlockValidation(t *testing.T) {
	keyPairs := make([]*accounts.PrivateKey, 10)
	for i := 0; i < len(keyPairs); i++ {
		privateKey, err := rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
		assert.NoError(t, err)
		privateKey.Add(privateKey, big.NewInt(1))
		keyPairs[i], err = accounts.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
		assert.NoError(t, err)
	}

	// Sort by x-coordinate of public key
	sort.Slice(keyPairs, func(i, j int) bool {
		return keyPairs[i].Pk.X.Cmp(&keyPairs[j].Pk.X) > 0
	})

	senders := make([]intMaxTypes.Sender, 128)
	for i, keyPair := range keyPairs {
		senders[i] = intMaxTypes.Sender{
			PublicKey: keyPair.Public(),
			AccountID: uint64(i) + 1,
			IsSigned:  true,
		}
	}

	defaultPublicKey := accounts.NewDummyPublicKey()
	for i := len(keyPairs); i < len(senders); i++ {
		senders[i] = intMaxTypes.Sender{
			PublicKey: defaultPublicKey,
			AccountID: 0,
			IsSigned:  false,
		}
	}

	txRoot, err := new(intMaxTypes.PoseidonHashOut).SetRandom()
	assert.NoError(t, err)

	senderPublicKeys := make([]byte, len(senders)*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senders {
		if sender.IsSigned {
			senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
		}
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
			assert.NoError(t, err)
			aggregatedSignature.Add(aggregatedSignature, signature)
		}
	}

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.AccountIDSenderType,
		senders,
		*txRoot,
		aggregatedSignature,
	)
	assert.NoError(t, blockContent.IsValid())

	_, err = intMaxTypes.MakePostRegistrationBlockInput(
		blockContent,
	)
	assert.NoError(t, err)

	// Usage of PostNonRegistrationBlock:
	// tx, err := intMaxTypes.PostNonRegistrationBlock(cfg, blockContent)
	// assert.NoError(t, err)

	// fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
}

func TestBlockHash(t *testing.T) {
	prevBlockHash := common.Hash{}
	depositRoot := common.Hash{}
	contentHash := common.Hash{}
	postedBlock := intMaxTypes.PostedBlock{
		PrevBlockHash: prevBlockHash,
		BlockNumber:   0,
		DepositRoot:   depositRoot,
		ContentHash:   contentHash,
	}

	currentHash := postedBlock.Hash()
	assert.Equal(t, currentHash.String(), "0x913fb9e1f6f1c6d910fd574a5cad8857aa43bfba24e401ada4f56090d4d997a7")
}
