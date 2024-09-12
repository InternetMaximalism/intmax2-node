package tree

import (
	"errors"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/types"
)

type TransferTree struct {
	Leaves []*types.Transfer
	inner  *PoseidonIncrementalMerkleTree
}

const TRANSFER_TREE_HEIGHT = 6

func NewTransferTree(
	height uint8,
	initialLeaves []*types.Transfer,
	zeroHash *PoseidonHashOut,
) (*TransferTree, error) {
	initialLeafHashes := make([]*PoseidonHashOut, len(initialLeaves))
	for key := range initialLeaves {
		initialLeafHashes[key] = initialLeaves[key].Hash()
	}

	t, err := NewPoseidonIncrementalMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, errors.Join(ErrNewPoseidonMerkleTreeFail, err)
	}

	leaves := make([]*types.Transfer, len(initialLeaves))
	for key := range initialLeaves {
		leaves[key] = new(types.Transfer).Set(initialLeaves[key])
	}

	return &TransferTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *TransferTree) BuildMerkleRoot(leaves []*PoseidonHashOut) (*PoseidonHashOut, error) {
	return t.inner.BuildMerkleRoot(leaves)
}

func (t *TransferTree) GetCurrentRootCountAndSiblings() (_ PoseidonHashOut, _ uint64, _ []*PoseidonHashOut) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *TransferTree) AddLeaf(index uint64, leaf *types.Transfer) (root *PoseidonHashOut, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return nil, errors.Join(ErrAddLeafFail, err)
	}

	if int(index) != len(t.Leaves) {
		return nil, errors.Join(ErrLeafInputIndexInvalid, errors.New("transfer tree AddLeaf"))
	}
	t.Leaves = append(t.Leaves, new(types.Transfer).Set(leaf))

	return root, nil
}

func (t *TransferTree) ComputeMerkleProof(
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
