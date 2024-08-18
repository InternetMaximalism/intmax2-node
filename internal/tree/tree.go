package tree

import (
	"errors"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/bits"
)

type PoseidonHashOut = goldenposeidon.PoseidonHashOut

type PoseidonMerkleTree struct {
	height     uint8
	zeroHashes []*PoseidonHashOut
	nodeHashes map[int]*PoseidonHashOut
}

// NewPoseidonMerkleTree creates new PoseidonMerkleTree by giving leaf nodes.
func NewPoseidonMerkleTree(
	height uint8,
	initialLeaves []*PoseidonHashOut,
	zeroHash *PoseidonHashOut,
) (mt *PoseidonMerkleTree, err error) {
	mt = &PoseidonMerkleTree{
		height:     height,
		zeroHashes: generateZeroHashes(height, zeroHash),
		nodeHashes: make(map[int]*PoseidonHashOut),
	}

	// TODO: Use initialLeaves.
	if len(initialLeaves) != 0 {
		panic("not implemented")
	}

	return mt, nil
}

func (t *PoseidonMerkleTree) GetRoot() *PoseidonHashOut {
	return t.GetNodeHash(1)
}

func (t *PoseidonMerkleTree) GetNodeHash(
	nodeIndex int,
) *PoseidonHashOut {
	if nodeIndex < 1 {
		panic("nodeIndex must be greater than 0")
	}
	if bits.Len(uint(nodeIndex))-1 > int(t.height) {
		panic("must be path.len() <= self.height")
	}

	if h, ok := t.nodeHashes[nodeIndex]; ok {
		return h
	}

	reversedIndex := len(t.zeroHashes) - bits.Len(uint(nodeIndex))

	return t.zeroHashes[reversedIndex]
}

func (t *PoseidonMerkleTree) getSiblingHash(nodeIndex int) *PoseidonHashOut {
	return t.GetNodeHash(nodeIndex ^ 1)
}

func (t *PoseidonMerkleTree) updateLeaf(
	index int,
	leafHash *PoseidonHashOut,
) {
	nodeIndex := 1<<int(t.height) + index

	h := new(PoseidonHashOut).Set(leafHash)
	t.nodeHashes[nodeIndex] = h

	for nodeIndex > 1 {
		sibling := t.getSiblingHash(nodeIndex)
		if nodeIndex&1 == 1 {
			h = goldenposeidon.Compress(sibling, h)
		} else {
			h = goldenposeidon.Compress(h, sibling)
		}
		nodeIndex >>= 1
		t.nodeHashes[nodeIndex] = h
	}
}

func (t *PoseidonMerkleTree) Prove(index int) ([]*PoseidonHashOut, error) {
	if index < 0 || index >= 1<<int(t.height) {
		var ErrMerkleTreeIndexOutOfRange = errors.New("the Merkle tree index out of range")
		return nil, ErrMerkleTreeIndexOutOfRange
	}

	nodeIndex := 1<<int(t.height) + index

	siblings := make([]*PoseidonHashOut, 0)
	for nodeIndex > 1 {
		siblings = append(siblings, t.getSiblingHash(nodeIndex))
		nodeIndex >>= 1
	}

	return siblings, nil
}
