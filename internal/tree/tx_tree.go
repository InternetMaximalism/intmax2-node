package tree

import (
	"errors"
	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type Tx struct {
	FeeTransferHash  *poseidonHashOut
	TransferTreeRoot *poseidonHashOut
}

func (t *Tx) Set(tx *Tx) *Tx {
	t.FeeTransferHash = tx.FeeTransferHash
	t.TransferTreeRoot = tx.TransferTreeRoot

	return t
}

func (t *Tx) SetRandom() (*Tx, error) {
	var err error
	t.FeeTransferHash, err = new(poseidonHashOut).SetRandom()
	if err != nil {
		return nil, err
	}
	t.TransferTreeRoot, err = new(poseidonHashOut).SetRandom()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Tx) ToFeSlice() []*ffg.Element {
	result := make([]*ffg.Element, 8)
	for i := 0; i < goldenposeidon.NUM_HASH_OUT_ELTS; i++ {
		result[i] = new(ffg.Element).Set(&t.FeeTransferHash.Elements[i])
	}
	for i := 0; i < goldenposeidon.NUM_HASH_OUT_ELTS; i++ {
		result[i+goldenposeidon.NUM_HASH_OUT_ELTS] = new(ffg.Element).Set(&t.TransferTreeRoot.Elements[i])
	}

	return result
}

func (t *Tx) Hash() *poseidonHashOut {
	input := t.ToFeSlice()
	return goldenposeidon.HashNoPad(input)
}

type TxTree struct {
	leaves []*Tx
	inner  *PoseidonMerkleTree
}

const TX_TREE_HEIGHT = 7

func NewTxTree(height uint8, initialLeaves []*Tx, zeroHash *poseidonHashOut) (*TxTree, error) {
	initialLeafHashes := make([]*poseidonHashOut, len(initialLeaves))
	for i, leaf := range initialLeaves {
		initialLeafHashes[i] = leaf.Hash()
	}

	t, err := NewPoseidonMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, err
	}

	leaves := make([]*Tx, len(initialLeaves))
	for i, leaf := range initialLeaves {
		leaves[i] = new(Tx).Set(leaf)
	}

	return &TxTree{
		leaves: leaves,
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

func (t *TxTree) AddLeaf(index uint64, leaf *Tx) (root *poseidonHashOut, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return nil, err
	}

	if int(index) != len(t.leaves) {
		return nil, errors.New("index is not equal to the length of leaves")
	}
	t.leaves = append(t.leaves, new(Tx).Set(leaf))

	return root, nil
}

func (t *TxTree) ComputeMerkleProof(index uint64, leaves []*poseidonHashOut) (siblings []*poseidonHashOut, root poseidonHashOut, err error) {
	return t.inner.ComputeMerkleProof(index, leaves)
}
