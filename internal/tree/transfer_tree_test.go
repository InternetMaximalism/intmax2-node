package tree

import (
	"crypto/rand"
	"fmt"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

const uint256Bits = 256

var maxUint256 = new(big.Int).Lsh(big.NewInt(1), uint256Bits)

func TestTransferData(t *testing.T) {
	// transferTree := transferTree{
	// 	Root: "0x5d9c1d8d9e1b2e1e1b2b1a8212d2430fbb7766c13d9ad305637dea3759065606",
	// }

	// transferTreeHash := transferTree.Hash()
	// assert.Equal(t, "0x5d9c1d8d9e1b2e1e1b2b1a8212d2430fbb7766c13d9ad305637dea3759065606", transferTreeHash.String())

	address := make([]byte, 32)
	_, err := rand.Read(address)
	assert.NoError(t, err)
	recipient := GenericAddress{
		addressType: EthereumAddressType,
		address:     address,
	}
	fmt.Printf("recipient: %+v\n", recipient.String())

	amount, err := rand.Int(rand.Reader, maxUint256)
	assert.NoError(t, err)
	salt, err := new(poseidonHashOut).SetRandom()
	assert.NoError(t, err)
	transferData := Transfer{
		Recipient:  recipient,
		TokenIndex: 0,
		Amount:     amount,
		Salt:       salt,
	}
	fmt.Printf("transfer data: %+v\n", transferData)

	flattenedTransfer := transferData.Marshal()
	fmt.Printf("flattened transfer: %+v\n", len(flattenedTransfer))

	transferHash := transferData.Hash()
	fmt.Printf("transfer hash: %+v\n", transferHash)
}

func TestTransferTree(t *testing.T) {
	transfers := make([]*Transfer, 8)

	for i := 0; i < 8; i++ {
		address := make([]byte, 32)
		_, err := rand.Read(address)
		assert.NoError(t, err)
		recipient := GenericAddress{
			addressType: EthereumAddressType,
			address:     address,
		}

		amount, err := rand.Int(rand.Reader, maxUint256)
		assert.NoError(t, err)
		salt, err := new(poseidonHashOut).SetRandom()
		assert.NoError(t, err)
		transferData := Transfer{
			Recipient:  recipient,
			TokenIndex: 0,
			Amount:     amount,
			Salt:       salt,
		}
		transfers[i] = &transferData
	}

	zeroHash := goldenposeidon.NewPoseidonHashOut()
	transferTree, err := NewTransferTree(3, transfers, zeroHash)
	assert.NoError(t, err)

	transferRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	expectedRoot := goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(transfers[0].Hash(), transfers[1].Hash()), goldenposeidon.Compress(transfers[2].Hash(), transfers[3].Hash())),
		goldenposeidon.Compress(goldenposeidon.Compress(transfers[4].Hash(), transfers[5].Hash()), goldenposeidon.Compress(transfers[6].Hash(), transfers[7].Hash())),
	)
	assert.Equal(t, expectedRoot.Elements, transferRoot.Elements)
}
