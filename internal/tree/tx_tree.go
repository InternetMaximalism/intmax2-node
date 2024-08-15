package tree

import (
	"errors"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/types"
)

type TxTree struct {
	Leaves []*types.Tx
	inner  *PoseidonIncrementalMerkleTree
}

const TX_TREE_HEIGHT = 7

func NewTxTree(height uint8, initialLeaves []*types.Tx, zeroHash *PoseidonHashOut) (*TxTree, error) {
	initialLeafHashes := make([]*PoseidonHashOut, len(initialLeaves))
	for i, leaf := range initialLeaves {
		initialLeafHashes[i] = leaf.Hash()
	}

	t, err := NewPoseidonIncrementalMerkleTree(height, initialLeafHashes, zeroHash)
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

func (t *TxTree) BuildMerkleRoot(leaves []*PoseidonHashOut) (root *PoseidonHashOut, err error) {
	return t.inner.BuildMerkleRoot(leaves)
}

// GetCurrentRootCountAndSiblings returns the latest root, count and sibblings
func (t *TxTree) GetCurrentRootCountAndSiblings() (root PoseidonHashOut, count uint64, siblings []*PoseidonHashOut) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *TxTree) AddLeaf(index uint64, leaf *types.Tx) (root *PoseidonHashOut, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return nil, errors.Join(ErrAddLeafFail, err)
	}

	if int(index) != len(t.Leaves) {
		return nil, ErrLeafInputIndexInvalid
	}
	t.Leaves = append(t.Leaves, new(types.Tx).Set(leaf))

	return root, nil
}

func (t *TxTree) ComputeMerkleProof(
	index uint64,
) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	leaves := make([]*goldenposeidon.PoseidonHashOut, 1<<t.inner.height)
	for i, leaf := range t.Leaves {
		leaves[i] = leaf.Hash()
	}
	for i := len(t.Leaves); i < len(leaves); i++ {
		leaves[i] = t.inner.zeroHashes[0]
	}

	return t.inner.ComputeMerkleProof(index, leaves)
}
