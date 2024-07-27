package tree

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

type BlockHashTree struct {
	Leaves [][32]byte
	inner  *KeccakMerkleTree
}

func NewBlockHashTree(height uint8, initialLeaves [][32]byte) (*BlockHashTree, error) {
	initialLeafHashes := make([][32]byte, len(initialLeaves))
	copy(initialLeafHashes, initialLeaves)

	t, err := NewKeccakMerkleTree(height, initialLeafHashes)
	if err != nil {
		return nil, err
	}

	leaves := make([][32]byte, len(initialLeaves))
	copy(leaves, initialLeaves)

	return &BlockHashTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *BlockHashTree) BuildMerkleRoot(leaves [][32]byte) (common.Hash, error) {
	return t.inner.BuildMerkleRoot(leaves)
}

func (t *BlockHashTree) GetCurrentRootCountAndSiblings() (common.Hash, uint32, [][32]byte) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *BlockHashTree) AddLeaf(index uint32, leaf [32]byte) (root [32]byte, err error) {
	leafHash := leaf
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return [32]byte{}, err
	}

	if int(index) != len(t.Leaves) {
		return [32]byte{}, errors.New("index is not equal to the length of leaves")
	}
	leaf = [32]byte{}
	copy(leaf[:], leafHash[:])
	t.Leaves = append(t.Leaves, leaf)

	return root, nil
}

func (t *BlockHashTree) ComputeMerkleProof(index uint32, leaves [][32]byte) (siblings [][32]byte, root common.Hash, err error) {
	return t.inner.ComputeMerkleProof(index, leaves)
}
