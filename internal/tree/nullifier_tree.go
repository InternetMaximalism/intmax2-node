package tree

import (
	intMaxTypes "intmax2-node/internal/types"
)

const NULLIFIER_TREE_HEIGHT uint8 = 32

// type NullifierLeaf struct {

type NullifierTree struct {
	inner *IndexedMerkleTree
}

func NewNullifierTree(height uint8) (*NullifierTree, error) {
	zeroHash := new(IndexedMerkleLeaf).EmptyLeaf().Hash()
	t, err := NewIndexedMerkleTree(height, zeroHash)
	if err != nil {
		return nil, err
	}

	return &NullifierTree{
		inner: t,
	}, nil
}

func (t *NullifierTree) Nullifiers() []intMaxTypes.Bytes32 {
	nullifiers := make([]intMaxTypes.Bytes32, len(t.inner.Leaves))
	for i := range t.inner.Leaves {
		key := new(intMaxTypes.Uint256).FromBigInt(t.inner.Leaves[i].Key)
		keyBytes := new(intMaxTypes.Bytes32).FromFieldElementSlice(key.ToFieldElementSlice())

		nullifiers[i] = *keyBytes
	}

	return nullifiers
}

func (t *NullifierTree) Set(other *NullifierTree) *NullifierTree {
	t.inner = new(IndexedMerkleTree).Set(other.inner)

	return t
}

func (t *NullifierTree) GetRoot() *PoseidonHashOut {
	root := t.inner.GetRoot()

	return root
}

func (t *NullifierTree) GetLeaf(index LeafIndex) *IndexedMerkleLeaf {
	return t.inner.Leaves[index]
}

func (t *NullifierTree) GetLeaves() []*IndexedMerkleLeaf {
	return t.inner.Leaves
}

func (t *NullifierTree) Prove(index LeafIndex) (proof *IndexedMerkleProof, root *PoseidonHashOut, err error) {
	return t.inner.Prove(index)
}

// func (t *NullifierTree) ProveMembership(key *big.Int) (membershipProof *IndexedMembershipProof, root *PoseidonHashOut, err error) {
// 	return t.inner.ProveMembership(key)
// }

func (t *NullifierTree) ProveMembership(key intMaxTypes.Bytes32) (membershipProof *IndexedMembershipProof, root *PoseidonHashOut, err error) {
	keyInt := new(intMaxTypes.Uint256).FromFieldElementSlice(key.ToFieldElementSlice())
	return t.inner.ProveMembership(keyInt.BigInt())
}

func (t *NullifierTree) Insert(key intMaxTypes.Bytes32) (proof *IndexedInsertionProof, err error) {
	keyInt := new(intMaxTypes.Uint256).FromFieldElementSlice(key.ToFieldElementSlice())
	return t.inner.Insert(keyInt.BigInt(), 0)
}

// func (t *NullifierTree) AddNullifiers(nullifiers []*intMaxTypes.Bytes32) error {
// 	for _, nullifier := range nullifiers {
// 		_, err := t.inner.Insert(nullifier)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
