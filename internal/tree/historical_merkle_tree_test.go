package tree_test

import (
	"fmt"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/tree"
	"testing"

	"github.com/stretchr/testify/require"
)

type poseidonHashOut = goldenposeidon.PoseidonHashOut

func TestNewHistoricalPoseidonMerkleTree(t *testing.T) {
	const cachingSubTreeHeight = 3
	zeroHash := new(poseidonHashOut).SetZero()
	storage := tree.NewMerkleTreeHistoryOnMemory()
	mt, err := tree.NewHistoricalPoseidonMerkleTree(32, zeroHash, storage, cachingSubTreeHeight)
	require.NoError(t, err)

	numLeaves := 530
	index := 261

	leaves := make([]*tree.PoseidonMerkleLeafWithIndex, numLeaves)
	for i := 0; i < len(leaves); i++ {
		leafHash, err := new(poseidonHashOut).SetRandom()
		require.NoError(t, err)
		leaves[i] = &tree.PoseidonMerkleLeafWithIndex{
			Index:    i,
			LeafHash: leafHash,
		}
	}

	_, err = mt.UpdateLeaves(leaves)
	require.NoError(t, err)

	root := mt.GetRoot()
	require.NoError(t, err)

	proof, err := mt.Prove(root, index)
	require.NoError(t, err)

	fmt.Printf("root = %s\n", root)
	fmt.Printf("leaves[%d] = %v\n", index, leaves[index].LeafHash)
	err = proof.Verify(root, index, leaves[index].LeafHash)
	require.NoError(t, err)

	// refresh the tree
	mt, err = tree.NewHistoricalPoseidonMerkleTree(32, zeroHash, storage, cachingSubTreeHeight)
	require.NoError(t, err)

	numLeaves = 260
	index = 261

	leaves2 := make([]*tree.PoseidonMerkleLeafWithIndex, numLeaves)
	for i := 0; i < len(leaves2); i++ {
		leafHash, err := new(poseidonHashOut).SetRandom()
		require.NoError(t, err)
		leaves2[i] = &tree.PoseidonMerkleLeafWithIndex{
			Index:    i,
			LeafHash: leafHash,
		}
	}

	_, err = mt.UpdateLeaves(leaves2)
	require.NoError(t, err)

	proof, err = mt.Prove(root, index)
	require.NoError(t, err)

	fmt.Printf("root = %s\n", root)
	fmt.Printf("leaves[%d] = %v\n", index, leaves[index].LeafHash)
	err = proof.Verify(root, index, leaves[index].LeafHash)
	require.NoError(t, err)

	root2 := mt.GetRoot()
	require.NoError(t, err)

	proof2, err := mt.Prove(root2, index)
	require.NoError(t, err)

	fmt.Printf("root2 = %s\n", root2)
	fmt.Printf("leaves[%d] = %v\n", index, leaves[index].LeafHash)
	err = proof2.Verify(root2, index, leaves[index].LeafHash)
	require.NoError(t, err)

	storage.ReportStats()
}
