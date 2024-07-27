package tree

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

type BlockHashTree struct {
	Leaves [][32]byte
	inner  *KeccakMerkleTree
}

func NewBlockHashTree(height uint8, initialLeaves [][numHashBytes]byte) (*BlockHashTree, error) {
	initialLeafHashes := make([][numHashBytes]byte, len(initialLeaves))
	copy(initialLeafHashes, initialLeaves)

	t, err := NewKeccakMerkleTree(height, initialLeafHashes)
	if err != nil {
		return nil, err
	}

	leaves := make([][numHashBytes]byte, len(initialLeaves))
	copy(leaves, initialLeaves)

	return &BlockHashTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *BlockHashTree) BuildMerkleRoot(leaves [][numHashBytes]byte) (common.Hash, error) {
	return t.inner.BuildMerkleRoot(leaves)
}

func (t *BlockHashTree) GetCurrentRootCountAndSiblings() (root common.Hash, nextIndex uint32, siblings [][numHashBytes]byte) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *BlockHashTree) AddLeaf(index uint32, leaf [numHashBytes]byte) (root [numHashBytes]byte, err error) {
	leafHash := leaf
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return [numHashBytes]byte{}, err
	}

	if int(index) != len(t.Leaves) {
		return [numHashBytes]byte{}, errors.New("index is not equal to the length of leaves")
	}
	leaf = [numHashBytes]byte{}
	copy(leaf[:], leafHash[:])
	t.Leaves = append(t.Leaves, leaf)

	return root, nil
}

func (t *BlockHashTree) ComputeMerkleProof(index uint32, leaves [][numHashBytes]byte) (siblings [][numHashBytes]byte, root common.Hash, err error) {
	return t.inner.ComputeMerkleProof(index, leaves)
}
