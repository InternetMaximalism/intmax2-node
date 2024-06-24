package tree

import (
	"errors"
	"intmax2-node/internal/types"
)

type TransferTree struct {
	Leaves []*types.Transfer
	inner  *PoseidonMerkleTree
}

func NewTransferTree(height uint8, initialLeaves []*types.Transfer, zeroHash *poseidonHashOut) (*TransferTree, error) {
	initialLeafHashes := make([]*poseidonHashOut, len(initialLeaves))
	for i, leaf := range initialLeaves {
		initialLeafHashes[i] = leaf.Hash()
	}

	t, err := NewPoseidonMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, err
	}

	leaves := make([]*types.Transfer, len(initialLeaves))
	for i, leaf := range initialLeaves {
		leaves[i] = new(types.Transfer).Set(leaf)
	}

	return &TransferTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *TransferTree) BuildMerkleRoot(leaves []*poseidonHashOut) (*poseidonHashOut, error) {
	return t.inner.BuildMerkleRoot(leaves)
}

func (t *TransferTree) GetCurrentRootCountAndSiblings() (poseidonHashOut, uint64, []*poseidonHashOut) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *TransferTree) AddLeaf(index uint64, leaf *types.Transfer) (root *poseidonHashOut, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return nil, err
	}

	if int(index) != len(t.Leaves) {
		return nil, errors.New("index is not equal to the length of leaves")
	}
	t.Leaves = append(t.Leaves, new(types.Transfer).Set(leaf))

	return root, nil
}

func (t *TransferTree) ComputeMerkleProof(index uint64, leaves []*poseidonHashOut) (siblings []*poseidonHashOut, root poseidonHashOut, err error) {
	return t.inner.ComputeMerkleProof(index, leaves)
}
