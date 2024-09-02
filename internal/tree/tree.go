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

func (t *PoseidonMerkleTree) Prove(index int) (*PoseidonMerkleProof, error) {
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

	return &PoseidonMerkleProof{
		Siblings: siblings,
	}, nil
}

type PoseidonMerkleProof struct {
	Siblings []*PoseidonHashOut
}

func (proof *PoseidonMerkleProof) GetMerkleRoot(
	index int,
	leafHash *PoseidonHashOut,
) *PoseidonHashOut {
	nodeHash := new(PoseidonHashOut).Set(leafHash)

	for _, sibling := range proof.Siblings {
		if index&1 == 1 {
			nodeHash = goldenposeidon.Compress(sibling, nodeHash)
		} else {
			nodeHash = goldenposeidon.Compress(nodeHash, sibling)
		}
		index >>= 1
	}

	return nodeHash
}

func (proof *PoseidonMerkleProof) Verify(
	root *PoseidonHashOut,
	index int,
	leafHash *PoseidonHashOut,
) error {
	expectedRoot := proof.GetMerkleRoot(index, leafHash)

	if !expectedRoot.Equal(root) {
		var ErrMerkleProofVerifyFail = errors.New("the Merkle proof verify fail")
		return ErrMerkleProofVerifyFail
	}

	return nil
}
