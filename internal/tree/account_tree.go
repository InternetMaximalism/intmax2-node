package tree

import (
	"math/big"
)

const ACCOUNT_TREE_HEIGHT uint8 = 40

type AccountTree struct {
	inner *IndexedMerkleTree
}

func NewAccountTree(height uint8) (*AccountTree, error) {
	zeroHash := new(PoseidonHashOut)
	t, err := NewIndexedMerkleTree(height, zeroHash)
	if err != nil {
		return nil, err
	}

	t.Update(big.NewInt(1), 0)

	return &AccountTree{
		inner: t,
	}, nil
}

func (t *AccountTree) GetRoot() PoseidonHashOut {
	root := t.inner.GetRoot()

	return root
}

func (t *AccountTree) GetLeaf(index uint64) *IndexedMerkleLeaf {
	return t.inner.GetLeaf(index)
}

func (t *AccountTree) Prove(index uint64) ([]*PoseidonHashOut, PoseidonHashOut, error) {
	return t.inner.Prove(index)
}
