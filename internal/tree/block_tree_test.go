package tree

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockTree(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	blockHashes := make([][32]byte, 8)

	for i := 0; i < 8; i++ {
		blockHash := [32]byte{}
		_, err := r.Read(blockHash[:])
		assert.NoError(t, err)
		blockHashes[i] = blockHash
	}

	blockHashTree, err := NewBlockHashTree(3, blockHashes)
	assert.NoError(t, err)

	blockHashTreeRoot, _, _ := blockHashTree.GetCurrentRootCountAndSiblings()

	expectedRoot := Hash(
		Hash(Hash(blockHashes[0], blockHashes[1]), Hash(blockHashes[2], blockHashes[3])),
		Hash(Hash(blockHashes[4], blockHashes[5]), Hash(blockHashes[6], blockHashes[7])),
	)
	assert.Equal(t, expectedRoot, ([32]byte)(blockHashTreeRoot))
}
