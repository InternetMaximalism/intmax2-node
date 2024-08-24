package tree_test

import (
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/tree"
	"testing"

	"github.com/stretchr/testify/require"
)

type poseidonHashOut = goldenposeidon.PoseidonHashOut

func TestNewHistoricalPoseidonMerkleTree(t *testing.T) {
	const cachingSubTreeHeight = 4
	zeroHash := new(poseidonHashOut).SetZero()
	mt, err := tree.NewHistoricalPoseidonMerkleTree(32, zeroHash, cachingSubTreeHeight)
	require.NoError(t, err)

	normalLeaves := []*poseidonHashOut{}

	numLeaves := 260
	index := 37

	for i := 0; i < numLeaves; i++ {
		e, err := new(poseidonHashOut).SetRandom()
		require.NoError(t, err)
		normalLeaves = append(normalLeaves, e)
	}

	leaves := make([]*tree.PoseidonMerkleLeafWithIndex, len(normalLeaves))
	for i := 0; i < len(normalLeaves); i++ {
		leaves[i] = &tree.PoseidonMerkleLeafWithIndex{
			Index:    i,
			LeafHash: normalLeaves[i],
		}
	}

	_, err = mt.UpdateLeaves(leaves)
	require.NoError(t, err)

	root := mt.GetRoot()
	require.NoError(t, err)

	proof, err := mt.Prove(root, index)
	require.NoError(t, err)

	err = proof.Verify(root, index, leaves[index].LeafHash)
	require.NoError(t, err)
}
