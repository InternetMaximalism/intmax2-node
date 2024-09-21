package tree

import (
	"errors"
	"fmt"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
)

const BLOCK_HASH_TREE_HEIGHT = 32

type BlockHashLeaf struct {
	leaf [numHashBytes]byte
}

func NewBlockHashLeaf(leaf [numHashBytes]byte) *BlockHashLeaf {
	h := new(BlockHashLeaf)
	copy(h.leaf[:], leaf[:])

	return h
}

func (leaf *BlockHashLeaf) SetDefault() *BlockHashLeaf {
	leaf.leaf = [numHashBytes]byte{}

	return leaf
}

func (leaf *BlockHashLeaf) Marshal() []byte {
	return leaf.leaf[:]
}

func (leaf *BlockHashLeaf) Hash() *PoseidonHashOut {
	b := intMaxTypes.Bytes32{}
	b.FromBytes(leaf.leaf[:])
	inputs := b.ToFieldElementSlice()

	return intMaxGP.HashNoPad(inputs)
}

type BlockHashMerkleProof = PoseidonMerkleProof

// type BlockHashTree struct {
// 	Leaves []*BlockHashLeaf
// 	inner  *PoseidonIncrementalMerkleTree
// }

// func (t *BlockHashTree) Set(other *BlockHashTree) *BlockHashTree {
// 	t.Leaves = make([]*BlockHashLeaf, len(other.Leaves))
// 	copy(t.Leaves, other.Leaves)
// 	t.inner = new(PoseidonIncrementalMerkleTree).Set(other.inner)

// 	return t
// }

// func NewBlockHashTreeWithInitialLeaves(height uint8, initialLeaves []*BlockHashLeaf) (*BlockHashTree, error) {
// 	initialLeafHashes := make([]*PoseidonHashOut, len(initialLeaves))
// 	for i, leaf := range initialLeaves {
// 		initialLeafHashes[i] = leaf.Hash()
// 	}

// 	zeroHash := new(BlockHashLeaf).SetDefault().Hash()
// 	t, err := NewPoseidonIncrementalMerkleTree(height, initialLeafHashes, zeroHash)
// 	if err != nil {
// 		return nil, err
// 	}

// 	leaves := make([]*BlockHashLeaf, len(initialLeaves))
// 	copy(leaves, initialLeaves)

// 	return &BlockHashTree{
// 		Leaves: leaves,
// 		inner:  t,
// 	}, nil
// }

// func (t *BlockHashTree) BuildMerkleRoot(leaves [][numHashBytes]byte) (*PoseidonHashOut, error) {
// 	initialLeaves := make([]*PoseidonHashOut, len(leaves))
// 	for i, leaf := range leaves {
// 		initialLeaves[i] = NewBlockHashLeaf(leaf).Hash()
// 	}

// 	return t.inner.BuildMerkleRoot(initialLeaves)
// }

// func (t *BlockHashTree) GetCurrentRootCountAndSiblings() (root PoseidonHashOut, nextIndex uint32, siblings []*PoseidonHashOut) {
// 	root, count, siblings := t.inner.GetCurrentRootCountAndSiblings()
// 	nextIndex = uint32(count)

// 	return root, nextIndex, siblings
// }

// func (t *BlockHashTree) AddLeaf(index uint32, leaf *BlockHashLeaf) (root *PoseidonHashOut, err error) {
// 	leafHash := leaf.Hash()
// 	root, err = t.inner.AddLeaf(uint64(index), leafHash)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if int(index) != len(t.Leaves) {
// 		return nil, errors.New("index is not equal to the length of block leaves")
// 	}

// 	t.Leaves = append(t.Leaves, leaf)

// 	return root, nil
// }

// func (t *BlockHashTree) ComputeMerkleProof(index uint32, leaves []BlockHashLeaf) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
// 	leafHashes := make([]*PoseidonHashOut, len(leaves))
// 	for i, leaf := range leaves {
// 		leafHashes[i] = leaf.Hash()
// 	}

// 	return t.inner.ComputeMerkleProof(uint64(index), leafHashes)
// }

// func (t *BlockHashTree) Prove(index uint32) (proof PoseidonMerkleProof, root PoseidonHashOut, err error) {
// 	leafHashes := make([]*PoseidonHashOut, len(t.Leaves))
// 	for i, leaf := range t.Leaves {
// 		leafHashes[i] = leaf.Hash()
// 	}

// 	siblings, root, err := t.inner.ComputeMerkleProof(uint64(index), leafHashes)
// 	if err != nil {
// 		return PoseidonMerkleProof{}, PoseidonHashOut{}, err
// 	}

// 	proof = PoseidonMerkleProof{
// 		Siblings: siblings,
// 	}

// 	// debug
// 	err = proof.Verify(leafHashes[index], int(index), &root)
// 	if err != nil {
// 		panic("proof.Verify failed")
// 	}

// 	return proof, root, err
// }

// func (t *BlockHashTree) GetRoot() *PoseidonHashOut {
// 	root, _, _ := t.inner.GetCurrentRootCountAndSiblings()

// 	return &root
// }

// func (t *BlockHashTree) GetLeaf(index uint32) *BlockHashLeaf {
// 	if int(index) >= len(t.Leaves) {
// 		return new(BlockHashLeaf).SetDefault()
// 	}

// 	return t.Leaves[index]
// }

type BlockHashTree struct {
	Leaves []*BlockHashLeaf
	inner  *PoseidonMerkleTree
}

func (t *BlockHashTree) Set(other *BlockHashTree) *BlockHashTree {
	t.Leaves = make([]*BlockHashLeaf, len(other.Leaves))
	copy(t.Leaves, other.Leaves)
	t.inner = new(PoseidonMerkleTree).Set(other.inner)

	return t
}

func NewBlockHashTreeWithInitialLeaves(height uint8, initialLeaves []*BlockHashLeaf) (*BlockHashTree, error) {
	zeroHash := new(BlockHashLeaf).SetDefault().Hash()
	t, err := NewPoseidonMerkleTree(height, zeroHash)
	if err != nil {
		return nil, err
	}

	for i, leaf := range initialLeaves {
		t.UpdateLeaf(i, leaf.Hash())
	}

	leaves := make([]*BlockHashLeaf, len(initialLeaves))
	copy(leaves, initialLeaves)

	return &BlockHashTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

// func (t *BlockHashTree) BuildMerkleRoot(leaves [][numHashBytes]byte) (*PoseidonHashOut, error) {
// 	initialLeaves := make([]*PoseidonHashOut, len(leaves))
// 	for i, leaf := range leaves {
// 		initialLeaves[i] = NewBlockHashLeaf(leaf).Hash()
// 	}

// 	return t.inner.BuildMerkleRoot(initialLeaves)
// }

func (t *BlockHashTree) GetCurrentRootCountAndSiblings() (root PoseidonHashOut, nextIndex uint32, siblings []*PoseidonHashOut) {
	// root, count, siblings := t.inner.GetCurrentRootCountAndSiblings()
	root = *t.inner.GetRoot()
	nextIndex = uint32(len(t.Leaves))
	proof, err := t.inner.Prove(int(nextIndex))
	if err != nil {
		panic(err)
	}

	return root, nextIndex, proof.Siblings
}

func (t *BlockHashTree) AddLeaf(index uint32, leaf *BlockHashLeaf) (root *PoseidonHashOut, err error) {
	leafHash := leaf.Hash()
	t.inner.UpdateLeaf(int(index), leafHash)

	if int(index) != len(t.Leaves) {
		return nil, errors.New("index is not equal to the length of block leaves")
	}

	root = t.inner.GetRoot()
	t.Leaves = append(t.Leaves, leaf)

	return root, nil
}

func (t *BlockHashTree) ComputeMerkleProof(index uint32, leaves []BlockHashLeaf) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	leafHashes := make([]*PoseidonHashOut, len(leaves))
	for i, leaf := range leaves {
		leafHashes[i] = leaf.Hash()
	}

	proof, err := t.inner.Prove(int(index))
	return proof.Siblings, *t.inner.GetRoot(), err
}

func (t *BlockHashTree) Prove(index uint32) (proof PoseidonMerkleProof, root *PoseidonHashOut, err error) {
	proof, err = t.inner.Prove(int(index))
	if err != nil {
		return proof, root, err
	}

	root = t.inner.GetRoot()
	leafHash := t.inner.GetLeaf(int(index))

	err = proof.Verify(leafHash, int(index), root)
	if err != nil {
		panic(fmt.Errorf("fatal error: Merkle proof verification failed: %w", err))
	}

	if !leafHash.Equal(t.GetLeaf(index).Hash()) {
		panic(fmt.Errorf("fatal error: leafHash is mismatch %s != %s", leafHash.String(), t.GetLeaf(index).Hash().String()))
	}

	return proof, root, err
}

func (t *BlockHashTree) GetRoot() *PoseidonHashOut {
	return t.inner.GetRoot()
}

func (t *BlockHashTree) GetLeaf(index uint32) *BlockHashLeaf {
	if int(index) >= len(t.Leaves) {
		return new(BlockHashLeaf).SetDefault()
	}

	return t.Leaves[index]
}
