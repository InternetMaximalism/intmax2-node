package tree

import (
	"errors"
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

type BlockHashMerkleProof = MerkleProof

type BlockHashTree struct {
	Leaves []*BlockHashLeaf
	inner  *PoseidonIncrementalMerkleTree
}

func NewBlockHashTreeWithInitialLeaves(height uint8, initialLeaves []*BlockHashLeaf) (*BlockHashTree, error) {
	initialLeafHashes := make([]*PoseidonHashOut, len(initialLeaves))
	for i, leaf := range initialLeaves {
		initialLeafHashes[i] = leaf.Hash()
	}

	zeroHash := new(BlockHashLeaf).SetDefault().Hash()
	t, err := NewPoseidonIncrementalMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, err
	}

	leaves := make([]*BlockHashLeaf, len(initialLeaves))
	copy(leaves, initialLeaves)

	return &BlockHashTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *BlockHashTree) BuildMerkleRoot(leaves [][numHashBytes]byte) (*PoseidonHashOut, error) {
	initialLeaves := make([]*PoseidonHashOut, len(leaves))
	for i, leaf := range leaves {
		initialLeaves[i] = NewBlockHashLeaf(leaf).Hash()
	}

	return t.inner.BuildMerkleRoot(initialLeaves)
}

func (t *BlockHashTree) GetCurrentRootCountAndSiblings() (root PoseidonHashOut, nextIndex uint32, siblings []*PoseidonHashOut) {
	root, count, siblings := t.inner.GetCurrentRootCountAndSiblings()
	nextIndex = uint32(count)

	return root, nextIndex, siblings
}

func (t *BlockHashTree) AddLeaf(index uint32, leaf *BlockHashLeaf) (root *PoseidonHashOut, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(uint64(index), leafHash)
	if err != nil {
		return nil, err
	}

	if int(index) != len(t.Leaves) {
		return nil, errors.New("index is not equal to the length of block leaves")
	}

	t.Leaves = append(t.Leaves, leaf)

	return root, nil
}

func (t *BlockHashTree) ComputeMerkleProof(index uint32, leaves []BlockHashLeaf) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	leafHashes := make([]*PoseidonHashOut, len(leaves))
	for i, leaf := range leaves {
		leafHashes[i] = leaf.Hash()
	}

	return t.inner.ComputeMerkleProof(uint64(index), leafHashes)
}

func (t *BlockHashTree) Prove(index uint32) (proof MerkleProof, root PoseidonHashOut, err error) {
	leafHashes := make([]*PoseidonHashOut, len(t.Leaves))
	for i, leaf := range t.Leaves {
		leafHashes[i] = leaf.Hash()
	}

	siblings, root, err := t.inner.ComputeMerkleProof(uint64(index), leafHashes)
	proof = MerkleProof{
		Siblings: siblings,
	}

	return proof, root, err
}

func (t *BlockHashTree) GetRoot() *PoseidonHashOut {
	root, _, _ := t.inner.GetCurrentRootCountAndSiblings()

	return &root
}
