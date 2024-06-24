package tree

import (
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestKeccakMerkleTreeWithoutInitialLeaves(t *testing.T) {
	zero := common.Hash{}
	initialLeaves := make([][32]byte, 0)
	mt, err := NewKeccakMerkleTree(3, initialLeaves)
	if err != nil {
		t.Errorf("fail to create merkle tree")
	}

	leaves := make([][32]byte, 8)
	for i := 0; i < 4; i++ {
		leaves[i] = [32]byte{}
		rand.Read(leaves[i][:])
		_, err := mt.AddLeaf(uint32(i), leaves[i])
		if err != nil {
			t.Errorf("fail to add leaf")
		}
	}

	expectedRoot := Hash(
		Hash(Hash(leaves[0], leaves[1]), Hash(leaves[2], leaves[3])),
		Hash(Hash(zero, zero), Hash(zero, zero)),
	)
	assert.Equal(t, expectedRoot, ([32]byte)(mt.currentRoot))

	leaves[4] = [32]byte{}
	rand.Read(leaves[4][:])
	_, err = mt.AddLeaf(4, leaves[4])
	if err != nil {
		t.Errorf("fail to add leaf")
	}

	expectedRoot = Hash(
		Hash(Hash(leaves[0], leaves[1]), Hash(leaves[2], leaves[3])),
		Hash(Hash(leaves[4], zero), Hash(zero, zero)),
	)
	assert.Equal(t, expectedRoot, ([32]byte)(mt.currentRoot))

	for i := 5; i < 8; i++ {
		leaves[i] = [32]byte{}
		rand.Read(leaves[i][:])
		_, err := mt.AddLeaf(uint32(i), leaves[i])
		if err != nil {
			t.Errorf("fail to add leaf")
		}
	}

	expectedRoot = Hash(
		Hash(Hash(leaves[0], leaves[1]), Hash(leaves[2], leaves[3])),
		Hash(Hash(leaves[4], leaves[5]), Hash(leaves[6], leaves[7])),
	)
	assert.Equal(t, expectedRoot, ([32]byte)(mt.currentRoot))
}
