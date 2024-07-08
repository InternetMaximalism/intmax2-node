package tree_test

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxTree(t *testing.T) {
	zeroTx := new(intMaxTypes.Tx).SetZero()
	zeroTxHash := zeroTx.Hash()
	initialLeaves := make([]*intMaxTypes.Tx, 0)
	mt, err := intMaxTree.NewTxTree(3, initialLeaves, zeroTxHash)
	require.Nil(t, err)

	leaves := make([]*intMaxTypes.Tx, 8)
	for i := 0; i < 4; i++ {
		leaves[i], err = intMaxTypes.NewTxWithPartialTransfers(uint64(i), 0xc1, []intMaxTypes.Transfer{})
		require.Nil(t, err)
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	expectedRoot := intMaxGP.Compress(
		intMaxGP.Compress(intMaxGP.Compress(leaves[0].Hash(), leaves[1].Hash()), intMaxGP.Compress(leaves[2].Hash(), leaves[3].Hash())),
		intMaxGP.Compress(intMaxGP.Compress(zeroTxHash, zeroTxHash), intMaxGP.Compress(zeroTxHash, zeroTxHash)),
	)
	// expectedRoot :=
	// 	intMaxGP.Compress(intMaxGP.Compress(leaves[0].Hash(), leaves[1].Hash()), intMaxGP.Compress(leaves[2].Hash(), leaves[3].Hash()))
	actualRoot, _, _ := mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)

	leaves[4], err = intMaxTypes.NewTxWithPartialTransfers(4, 0xc1, []intMaxTypes.Transfer{})
	assert.Nil(t, err)
	_, err = mt.AddLeaf(4, leaves[4])
	require.Nil(t, err)

	expectedRoot = intMaxGP.Compress(
		intMaxGP.Compress(intMaxGP.Compress(leaves[0].Hash(), leaves[1].Hash()), intMaxGP.Compress(leaves[2].Hash(), leaves[3].Hash())),
		intMaxGP.Compress(intMaxGP.Compress(leaves[4].Hash(), zeroTxHash), intMaxGP.Compress(zeroTxHash, zeroTxHash)),
	)
	actualRoot, _, _ = mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)

	for i := 5; i < 8; i++ {
		leaves[i], err = intMaxTypes.NewTxWithPartialTransfers(uint64(i), 0xc1, []intMaxTypes.Transfer{})
		assert.Nil(t, err)
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	expectedRoot = intMaxGP.Compress(
		intMaxGP.Compress(intMaxGP.Compress(leaves[0].Hash(), leaves[1].Hash()), intMaxGP.Compress(leaves[2].Hash(), leaves[3].Hash())),
		intMaxGP.Compress(intMaxGP.Compress(leaves[4].Hash(), leaves[5].Hash()), intMaxGP.Compress(leaves[6].Hash(), leaves[7].Hash())),
	)
	actualRoot, _, _ = mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)
}
