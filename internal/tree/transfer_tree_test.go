package tree_test

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxtypes "intmax2-node/internal/types"
	"math/big"
	"math/rand"
	"testing"

	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
)

const uint256Bits = 256

var maxUint256 = new(big.Int).Lsh(big.NewInt(1), uint256Bits)

func TestTransferTree(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	transfers := make([]*intMaxtypes.Transfer, 8)

	for i := 0; i < 8; i++ {
		address := make([]byte, 32)
		_, err := r.Read(address)
		assert.NoError(t, err)
		recipient, err := intMaxtypes.NewINTMAXAddress(address)
		assert.NoError(t, err)
		assert.NotNil(t, recipient)
		amount := new(big.Int).Rand(r, maxUint256)
		assert.NoError(t, err)
		salt := new(intMaxTree.PoseidonHashOut)
		salt.Elements[0] = *new(ffg.Element).SetUint64(1)
		salt.Elements[1] = *new(ffg.Element).SetUint64(2)
		salt.Elements[2] = *new(ffg.Element).SetUint64(3)
		salt.Elements[3] = *new(ffg.Element).SetUint64(4)
		transferData := intMaxtypes.Transfer{
			Recipient:  recipient,
			TokenIndex: 0,
			Amount:     amount,
			Salt:       salt,
		}
		transfers[i] = &transferData
	}

	zeroHash := intMaxGP.NewPoseidonHashOut()
	transferTree, err := intMaxTree.NewTransferTree(3, transfers, zeroHash)
	assert.NoError(t, err)

	transferRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	expectedRoot := intMaxGP.Compress(
		intMaxGP.Compress(intMaxGP.Compress(transfers[0].Hash(), transfers[1].Hash()), intMaxGP.Compress(transfers[2].Hash(), transfers[3].Hash())),
		intMaxGP.Compress(intMaxGP.Compress(transfers[4].Hash(), transfers[5].Hash()), intMaxGP.Compress(transfers[6].Hash(), transfers[7].Hash())),
	)
	assert.Equal(t, expectedRoot.Elements, transferRoot.Elements)
}
