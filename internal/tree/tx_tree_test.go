package tree

import (
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxTree(t *testing.T) {
	zeroTx := types.Tx{
		FeeTransferHash:  goldenposeidon.NewPoseidonHashOut(),
		TransferTreeRoot: goldenposeidon.NewPoseidonHashOut(),
	}
	zeroTxHash := zeroTx.Hash()
	initialLeaves := make([]*types.Tx, 0)
	mt, err := NewTxTree(3, initialLeaves, zeroTxHash)
	require.Nil(t, err)

	leaves := make([]*types.Tx, 8)
	for i := 0; i < 4; i++ {
		leaves[i], _ = new(types.Tx).SetRandom()
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	expectedRoot := goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0].Hash(), leaves[1].Hash()), goldenposeidon.Compress(leaves[2].Hash(), leaves[3].Hash())),
		goldenposeidon.Compress(goldenposeidon.Compress(zeroTxHash, zeroTxHash), goldenposeidon.Compress(zeroTxHash, zeroTxHash)),
	)
	// expectedRoot :=
	// 	goldenposeidon.Compress(goldenposeidon.Compress(leaves[0].Hash(), leaves[1].Hash()), goldenposeidon.Compress(leaves[2].Hash(), leaves[3].Hash()))
	actualRoot, _, _ := mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)

	leaves[4], _ = new(types.Tx).SetRandom()
	_, err = mt.AddLeaf(4, leaves[4])
	require.Nil(t, err)

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0].Hash(), leaves[1].Hash()), goldenposeidon.Compress(leaves[2].Hash(), leaves[3].Hash())),
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[4].Hash(), zeroTxHash), goldenposeidon.Compress(zeroTxHash, zeroTxHash)),
	)
	actualRoot, _, _ = mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)

	for i := 5; i < 8; i++ {
		leaves[i], _ = new(types.Tx).SetRandom()
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0].Hash(), leaves[1].Hash()), goldenposeidon.Compress(leaves[2].Hash(), leaves[3].Hash())),
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[4].Hash(), leaves[5].Hash()), goldenposeidon.Compress(leaves[6].Hash(), leaves[7].Hash())),
	)
	actualRoot, _, _ = mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)
}
