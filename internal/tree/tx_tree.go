package tree

import (
	"errors"
	"intmax2-node/internal/types"
)

type TxTree struct {
	Leaves []*types.Tx
	inner  *PoseidonMerkleTree
}

const TX_TREE_HEIGHT = 7

func NewTxTree(height uint8, initialLeaves []*types.Tx, zeroHash *poseidonHashOut) (*TxTree, error) {
	initialLeafHashes := make([]*poseidonHashOut, len(initialLeaves))
	for i, leaf := range initialLeaves {
		initialLeafHashes[i] = leaf.Hash()
	}

	t, err := NewPoseidonMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, err
	}

	leaves := make([]*types.Tx, len(initialLeaves))
	for i, leaf := range initialLeaves {
		leaves[i] = new(types.Tx).Set(leaf)
	}

	return &TxTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *TxTree) BuildMerkleRoot(leaves []*poseidonHashOut) (root *poseidonHashOut, err error) {
	return t.inner.BuildMerkleRoot(leaves)
}

// GetCurrentRootCountAndSiblings returns the latest root, count and sibblings
func (t *TxTree) GetCurrentRootCountAndSiblings() (root poseidonHashOut, count uint64, siblings []*poseidonHashOut) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *TxTree) AddLeaf(index uint64, leaf *types.Tx) (root *poseidonHashOut, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return nil, err
	}

	if int(index) != len(t.Leaves) {
		return nil, errors.New("index is not equal to the length of leaves")
	}
	t.Leaves = append(t.Leaves, new(types.Tx).Set(leaf))

	return root, nil
}

func (t *TxTree) ComputeMerkleProof(index uint64, leaves []*poseidonHashOut) (siblings []*poseidonHashOut, root poseidonHashOut, err error) {
	return t.inner.ComputeMerkleProof(index, leaves)
}
