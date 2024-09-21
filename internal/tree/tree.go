package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/bits"
)

type PoseidonHashOut = goldenposeidon.PoseidonHashOut

type PoseidonMerkleTree struct {
	height     uint8
	zeroHashes []*PoseidonHashOut
	nodeHashes map[int]*PoseidonHashOut
}

func (t *PoseidonMerkleTree) Set(other *PoseidonMerkleTree) *PoseidonMerkleTree {
	t.height = other.height
	t.zeroHashes = make([]*PoseidonHashOut, len(other.zeroHashes))
	copy(t.zeroHashes, other.zeroHashes)
	t.nodeHashes = make(map[int]*PoseidonHashOut)
	for k, v := range other.nodeHashes {
		t.nodeHashes[k] = v
	}

	return t
}

func (t *PoseidonMerkleTree) ClearCache() {
	t.nodeHashes = make(map[int]*PoseidonHashOut)
}

// NewPoseidonMerkleTree creates new PoseidonMerkleTree by giving leaf nodes.
func NewPoseidonMerkleTree(
	height uint8,
	zeroHash *PoseidonHashOut,
) (mt *PoseidonMerkleTree, err error) {
	mt = &PoseidonMerkleTree{
		height:     height,
		zeroHashes: generateZeroHashes(height, zeroHash),
		nodeHashes: make(map[int]*PoseidonHashOut),
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

	return t.getZeroHash(nodeIndex)
}

func (t *PoseidonMerkleTree) getZeroHash(nodeIndex int) *PoseidonHashOut {
	reversedIndex := len(t.zeroHashes) - bits.Len(uint(nodeIndex))

	return t.zeroHashes[reversedIndex]
}

func (t *PoseidonMerkleTree) getSiblingHash(nodeIndex int) *PoseidonHashOut {
	return t.GetNodeHash(nodeIndex ^ 1)
}

func (t *PoseidonMerkleTree) UpdateLeaf(
	index int,
	leafHash *PoseidonHashOut,
) {
	t.updateLeaf(index, leafHash)
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

func (t *PoseidonMerkleTree) Prove(index int) (proof PoseidonMerkleProof, err error) {
	if index < 0 || index >= 1<<int(t.height) {
		var ErrMerkleTreeIndexOutOfRange = errors.New("the Merkle tree index out of range")
		return proof, ErrMerkleTreeIndexOutOfRange
	}

	nodeIndex := 1<<int(t.height) + index
	leafHash := t.GetNodeHash(nodeIndex)

	siblings := make([]*PoseidonHashOut, 0)
	for nodeIndex > 1 {
		siblings = append(siblings, t.getSiblingHash(nodeIndex))
		nodeIndex >>= 1
	}

	proof = PoseidonMerkleProof{
		Siblings: siblings,
	}

	root := t.GetRoot()
	err = proof.Verify(leafHash, index, root)
	if err != nil {
		panic("MerkleProof proof.Verify failed")
	}

	return proof, nil
}

func (t *PoseidonMerkleTree) ProveWithLeaf(index int) (PoseidonMerkleProof, *PoseidonHashOut, *PoseidonHashOut, error) {
	proof, err := t.Prove(index)
	if err != nil {
		return PoseidonMerkleProof{}, nil, nil, err
	}

	leaf := t.GetLeaf(index)
	nodeIndex := 1<<int(t.height) + index
	leafHash := t.GetNodeHash(nodeIndex)
	if !leafHash.Equal(leaf) {
		panic("leafHash != leaf")
	}
	root := t.GetRoot()

	return proof, leaf, root, err
}

func (t *PoseidonMerkleTree) GetLeaf(index int) *PoseidonHashOut {
	nodeIndex := 1<<int(t.height) + index

	return t.GetNodeHash(nodeIndex)
}

type PoseidonMerkleProof struct {
	Siblings []*goldenposeidon.PoseidonHashOut
}

func (proof *PoseidonMerkleProof) Set(other *PoseidonMerkleProof) *PoseidonMerkleProof {
	proof.Siblings = make([]*goldenposeidon.PoseidonHashOut, len(other.Siblings))
	copy(proof.Siblings, other.Siblings)
	return proof
}

func (proof *PoseidonMerkleProof) MarshalJSON() ([]byte, error) {
	return json.Marshal(proof.Siblings)
}

func (proof *PoseidonMerkleProof) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &proof.Siblings)
}

func (proof *PoseidonMerkleProof) GetRoot(leafHash *goldenposeidon.PoseidonHashOut, index int) *goldenposeidon.PoseidonHashOut {
	height := len(proof.Siblings)
	if index >= 1<<uint(height) {
		panic("index out of bounds")
	}
	nodeIndex := 1<<uint(height) + index
	h := new(PoseidonHashOut).Set(leafHash)

	for i := 0; i < height; i++ {
		sibling := proof.Siblings[i]
		if nodeIndex&1 == 1 {
			h = goldenposeidon.Compress(sibling, h)
		} else {
			h = goldenposeidon.Compress(h, sibling)
		}
		nodeIndex >>= 1
	}
	if nodeIndex != 1 {
		panic("invalid nodeIndex")
	}

	return h
}

func (proof *PoseidonMerkleProof) Verify(leafHash *goldenposeidon.PoseidonHashOut, index int, root *goldenposeidon.PoseidonHashOut) error {
	computedRoot := proof.GetRoot(leafHash, index)
	if !computedRoot.Equal(root) {
		return fmt.Errorf("invalid root: %v != %v", computedRoot, root)
	}

	return nil
}
