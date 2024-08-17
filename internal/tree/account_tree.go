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
	return t.inner.GetRoot()
}

func (t *AccountTree) GetLeaf(accountID uint64) *IndexedMerkleLeaf {
	return t.inner.GetLeaf(LeafIndex(accountID))
}

func (t *AccountTree) GetAccountID(publicKey *big.Int) (accountID uint64, ok bool) {
	index, ok := t.inner.GetIndex(publicKey)

	return uint64(index), ok
}

func (t *AccountTree) Prove(accountID int) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	return t.inner.Prove(LeafIndex(accountID)) // nolint:unconvert
}

func (t *AccountTree) ProveMembership(publicKey *big.Int) (proof *IndexedMembershipProof, root PoseidonHashOut, err error) {
	return t.inner.ProveMembership(publicKey)
}

func (t *AccountTree) Insert(publicKey *big.Int, lastSentBlockNumber uint64) (*IndexedInsertionProof, error) {
	return t.inner.Insert(publicKey, lastSentBlockNumber)
}

func (t *AccountTree) Update(publicKey *big.Int, lastSentBlockNumber uint64) (*IndexedUpdateProof, error) {
	return t.inner.Update(publicKey, lastSentBlockNumber)
}
