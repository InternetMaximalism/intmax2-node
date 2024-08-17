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

	_, err = t.Insert(big.NewInt(1), 0)
	if err != nil {
		return nil, err
	}

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

func (t *AccountTree) Prove(index uint64) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	return t.inner.Prove(index)
}

func (t *AccountTree) ProveMembership(key *big.Int) (proof *IndexedMembershipProof, root PoseidonHashOut, err error) {
	return t.inner.ProveMembership(key)
}
