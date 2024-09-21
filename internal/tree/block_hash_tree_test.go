package tree_test

import (
	"errors"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	tree "intmax2-node/internal/tree"
)

const numHashBytes = 32
const BLOCK_HASH_TREE_HEIGHT = 32

type PoseidonHashOut = intMaxGP.PoseidonHashOut

type BlockHashMerkleProof = tree.MerkleProof

type BlockHashTree struct {
	Leaves          []*tree.BlockHashLeaf
	NextBlockNumber uint32
	inner           *tree.PoseidonMerkleTree
}

func (t *BlockHashTree) Set(other *BlockHashTree) *BlockHashTree {
	t.Leaves = make([]*tree.BlockHashLeaf, len(other.Leaves))
	copy(t.Leaves, other.Leaves)
	t.inner = new(tree.PoseidonMerkleTree).Set(other.inner)

	return t
}

func NewBlockHashTreeWithInitialLeaves(height uint8, initialLeaves []*tree.BlockHashLeaf) (*BlockHashTree, error) {
	zeroHash := new(tree.BlockHashLeaf).SetDefault().Hash()
	t, err := tree.NewPoseidonMerkleTree(height, zeroHash)
	if err != nil {
		return nil, err
	}

	for i, leaf := range initialLeaves {
		t.UpdateLeaf(i, leaf.Hash())
	}

	leaves := make([]*tree.BlockHashLeaf, len(initialLeaves))
	copy(leaves, initialLeaves)

	return &BlockHashTree{
		Leaves:          leaves,
		NextBlockNumber: uint32(len(initialLeaves)),
		inner:           t,
	}, nil
}

// func (t *BlockHashTree) BuildMerkleRoot(leaves [][numHashBytes]byte) (*PoseidonHashOut, error) {
// 	initialLeaves := make([]*PoseidonHashOut, len(leaves))
// 	for i, leaf := range leaves {
// 		initialLeaves[i] = NewBlockHashLeaf(leaf).Hash()
// 	}

// 	return t.inner.BuildMerkleRoot(initialLeaves)
// }

func (t *BlockHashTree) GetCurrentRootCountAndSiblings() (root tree.PoseidonHashOut, nextIndex uint32, siblings []*tree.PoseidonHashOut) {
	// root, count, siblings := t.inner.GetCurrentRootCountAndSiblings()
	root = *t.inner.GetRoot()
	nextIndex = t.NextBlockNumber
	proof, err := t.inner.Prove(int(nextIndex))
	if err != nil {
		panic(err)
	}

	return root, nextIndex, proof.Siblings
}

func (t *BlockHashTree) AddLeaf(index uint32, leaf *tree.BlockHashLeaf) (root *PoseidonHashOut, err error) {
	leafHash := leaf.Hash()
	t.inner.UpdateLeaf(int(index), leafHash)
	// if err != nil {
	// 	return nil, err
	// }

	if int(index) != len(t.Leaves) {
		return nil, errors.New("index is not equal to the length of block leaves")
	}

	t.Leaves = append(t.Leaves, leaf)
	t.NextBlockNumber++

	return root, nil
}

func (t *BlockHashTree) ComputeMerkleProof(index uint32, leaves []tree.BlockHashLeaf) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	leafHashes := make([]*PoseidonHashOut, len(leaves))
	for i, leaf := range leaves {
		leafHashes[i] = leaf.Hash()
	}

	proof, err := t.inner.Prove(int(index))
	return proof.Siblings, *t.inner.GetRoot(), err
}

func (t *BlockHashTree) Prove(index uint32) (proof tree.MerkleProof, root PoseidonHashOut, err error) {
	leafHashes := make([]*PoseidonHashOut, len(t.Leaves))
	for i, leaf := range t.Leaves {
		leafHashes[i] = leaf.Hash()
	}

	merkleProof, err := t.inner.Prove(int(index))
	if err != nil {
		return tree.MerkleProof{}, PoseidonHashOut{}, err
	}

	proof = tree.MerkleProof{Siblings: merkleProof.Siblings}

	// debug
	err = proof.Verify(&root, int(index), leafHashes[index])
	if err != nil {
		panic("TEST proof.Verify failed")
	}

	return proof, *t.inner.GetRoot(), err
}

func (t *BlockHashTree) GetRoot() *PoseidonHashOut {
	return t.inner.GetRoot()
}

func (t *BlockHashTree) GetLeaf(index uint32) *tree.BlockHashLeaf {
	if int(index) >= len(t.Leaves) {
		return new(tree.BlockHashLeaf).SetDefault()
	}

	return t.Leaves[index]
}
