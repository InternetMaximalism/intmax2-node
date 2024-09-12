package tree

import (
	"math/big"
)

const ACCOUNT_TREE_HEIGHT uint8 = 40

type AccountTree struct {
	inner *IndexedMerkleTree
}

func (t *AccountTree) Set(other *AccountTree) *AccountTree {
	t.inner = new(IndexedMerkleTree).Set(other.inner)

	return t
}

func NewAccountTree(height uint8) (*AccountTree, error) {
	zeroHash := new(IndexedMerkleLeaf).EmptyLeaf().Hash()
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

func (t *AccountTree) GetRoot() *PoseidonHashOut {
	return t.inner.GetRoot()
}

func (t *AccountTree) GetLeaf(accountID uint64) *IndexedMerkleLeaf {
	return t.inner.GetLeaf(LeafIndex(accountID))
}

func (t *AccountTree) Count() int {
	return len(t.inner.Leaves)
}

func (t *AccountTree) GetAccountID(publicKey *big.Int) (accountID uint64, ok bool) {
	index, ok := t.inner.GetIndex(publicKey)

	return uint64(index), ok
}

func (t *AccountTree) Prove(accountID uint64) (proof *IndexedMerkleProof, root *PoseidonHashOut, err error) {
	return t.inner.Prove(LeafIndex(accountID))
}

func (t *AccountTree) ProveMembership(publicKey *big.Int) (proof *IndexedMembershipProof, root *PoseidonHashOut, err error) {
	return t.inner.ProveMembership(publicKey)
}

func (t *AccountTree) Insert(publicKey *big.Int, lastSentBlockNumber uint64) (*IndexedInsertionProof, error) {
	return t.inner.Insert(publicKey, lastSentBlockNumber)
}

func (t *AccountTree) Update(publicKey *big.Int, lastSentBlockNumber uint64) (*IndexedUpdateProof, error) {
	return t.inner.Update(publicKey, lastSentBlockNumber)
}

func NewDummyAccountRegistrationProof(height uint8) *IndexedInsertionProof {
	return &IndexedInsertionProof{
		Index:        0,
		LowLeafProof: NewDummyIndexedMerkleProof(height),
		LeafProof:    NewDummyIndexedMerkleProof(height),
		LowLeafIndex: 0,
		PrevLowLeaf:  new(IndexedMerkleLeaf).SetDefault(),
	}
}
