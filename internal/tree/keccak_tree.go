package tree

/// This is a modification of the code from the following URL:
/// https://github.com/0xPolygonHermez/zkevm-node/blob/develop/l1infotree/hash.go

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"golang.org/x/crypto/sha3"
)

const (
	int2Key      = 2
	numHashBytes = 32
)

// KeccakMerkleTree provides methods to compute KeccakMerkleTree
type KeccakMerkleTree struct {
	height      uint8
	zeroHashes  [][32]byte
	count       uint32
	siblings    [][32]byte
	currentRoot common.Hash
}

// NewKeccakMerkleTree creates new KeccakMerkleTree.
func NewKeccakMerkleTree(height uint8, initialLeaves [][numHashBytes]byte, zeroHash common.Hash) (*KeccakMerkleTree, error) {
	mt := &KeccakMerkleTree{
		zeroHashes: generateKeccakZeroHashes(height, zeroHash),
		height:     height,
		count:      uint32(len(initialLeaves)),
	}
	var err error
	var siblings *KeccakMerkleProof
	siblings, mt.currentRoot, err = mt.initSiblings(initialLeaves)
	if err != nil {
		log.Error("error initializing si siblings. Error: ", err)
		return nil, err
	}
	log.Debug("Initial count: ", mt.count)
	log.Debug("Initial root: ", mt.currentRoot)
	mt.siblings = siblings.Siblings
	return mt, nil
}

func buildKeccakIntermediate(leaves [][numHashBytes]byte) (_ [][][]byte, _ [][numHashBytes]byte) {
	var (
		nodes  [][][]byte
		hashes [][numHashBytes]byte
	)
	for i := 0; i < len(leaves); i += 2 {
		left := i
		right := i + 1
		hash := Hash(leaves[left], leaves[right])
		nodes = append(nodes, [][]byte{hash[:], leaves[left][:], leaves[right][:]})
		hashes = append(hashes, hash)
	}
	return nodes, hashes
}

// BuildKeccakRoot computes the root given the leaves of the tree
func (mt *KeccakMerkleTree) BuildMerkleRoot(leaves [][numHashBytes]byte) (common.Hash, error) {
	var (
		nodes [][][][]byte
		ns    [][][]byte
	)
	if len(leaves) == 0 {
		leaves = append(leaves, mt.zeroHashes[0])
	}

	for h := uint8(0); h < mt.height; h++ {
		if len(leaves)%2 == 1 {
			leaves = append(leaves, mt.zeroHashes[h])
		}
		ns, leaves = buildKeccakIntermediate(leaves)
		nodes = append(nodes, ns)
	}
	if len(ns) != 1 {
		return common.Hash{}, fmt.Errorf("error: more than one root detected: %+v", nodes)
	}

	return common.BytesToHash(ns[0][0]), nil
}

// ComputeMerkleProof computes the merkleProof and root given the leaves of the tree
func (mt *KeccakMerkleTree) ComputeMerkleProof(gerIndex uint32, leaves [][numHashBytes]byte) (*KeccakMerkleProof, common.Hash, error) {
	var ns [][][]byte
	if len(leaves) == 0 {
		leaves = append(leaves, mt.zeroHashes[0])
	}
	var siblings [][numHashBytes]byte
	index := gerIndex
	for h := uint8(0); h < mt.height; h++ {
		if len(leaves)%2 == 1 {
			leaves = append(leaves, mt.zeroHashes[h])
		}
		if index%2 == 1 {
			siblings = append(siblings, leaves[index-1])
		} else if len(leaves) > 1 {
			if index >= uint32(len(leaves)) {
				siblings = append(siblings, mt.zeroHashes[h])
			} else {
				siblings = append(siblings, leaves[index+1])
			}
		}
		var (
			nsi    [][][]byte
			hashes [][numHashBytes]byte
		)
		for i := 0; i < len(leaves); i += int2Key {
			left := i
			right := i + 1
			hash := Hash(leaves[left], leaves[right])
			nsi = append(nsi, [][]byte{hash[:], leaves[left][:], leaves[right][:]})
			hashes = append(hashes, hash)
		}
		// Find the index of the leave in the next level of the tree.
		// Divide the index by 2 to find the position in the upper level
		index = uint32(float64(index) / int2Key) //nolint:gomnd
		ns = nsi
		leaves = hashes
	}
	if len(ns) != 1 {
		return nil, common.Hash{}, fmt.Errorf("error: more than one root detected: %+v", ns)
	}

	return &KeccakMerkleProof{Siblings: siblings}, common.BytesToHash(ns[0][0]), nil
}

// AddLeaf adds new leaves to the tree and computes the new root
func (mt *KeccakMerkleTree) AddLeaf(index uint32, leaf [numHashBytes]byte) (common.Hash, error) {
	if index != mt.count {
		return common.Hash{}, fmt.Errorf("mismatched leaf count: %d, expected: %d", index, mt.count)
	}
	cur := leaf
	isFilledSubTree := true

	for h := uint8(0); h < mt.height; h++ {
		if index&(1<<h) > 0 {
			var child [numHashBytes]byte
			copy(child[:], cur[:])
			parent := Hash(mt.siblings[h], child)
			cur = parent
		} else {
			if isFilledSubTree {
				// we will update the sibling when the sub tree is complete
				copy(mt.siblings[h][:], cur[:])
				// we have a left child in this layer, it means the right child is empty so the sub tree is not completed
				isFilledSubTree = false
			}
			var child [numHashBytes]byte
			copy(child[:], cur[:])
			parent := Hash(child, mt.zeroHashes[h])
			cur = parent
			// the sibling of 0 bit should be the zero hash, since we are in the last node of the tree
		}
	}
	mt.currentRoot = cur
	mt.count++
	return cur, nil
}

// initSiblings returns the siblings of the node at the given index.
// it is used to initialize the siblings array in the beginning.
func (mt *KeccakMerkleTree) initSiblings(initialLeaves [][numHashBytes]byte) (*KeccakMerkleProof, common.Hash, error) {
	if mt.count != uint32(len(initialLeaves)) {
		return nil, [numHashBytes]byte{}, fmt.Errorf("error: mt.count and initialLeaves length mismatch")
	}
	if mt.count == 0 {
		var siblings [][numHashBytes]byte
		for h := 0; h < int(mt.height); h++ {
			var left [numHashBytes]byte
			copy(left[:], mt.zeroHashes[h][:])
			siblings = append(siblings, left)
		}
		root, err := mt.BuildMerkleRoot(initialLeaves)
		if err != nil {
			log.Error("error calculating initial root: ", err)
			return nil, [numHashBytes]byte{}, err
		}
		return &KeccakMerkleProof{Siblings: siblings}, root, nil
	}

	return mt.ComputeMerkleProof(mt.count, initialLeaves)
}

// GetCurrentRootCountAndSiblings returns the latest root, count and sibblings
func (mt *KeccakMerkleTree) GetCurrentRootCountAndSiblings() (root common.Hash, nextIndex uint32, siblings *KeccakMerkleProof) {
	siblings = &KeccakMerkleProof{
		Siblings: mt.siblings,
	}

	return mt.currentRoot, mt.count, siblings
}

// Hash calculates the keccak hash of elements.
func Hash(data ...[numHashBytes]byte) [numHashBytes]byte {
	var res [32]byte
	hash := sha3.NewLegacyKeccak256()
	for _, d := range data {
		hash.Write(d[:]) //nolint:errcheck,gosec
	}
	copy(res[:], hash.Sum(nil))
	return res
}

func generateKeccakZeroHashes(height uint8, initialZeroHash common.Hash) [][numHashBytes]byte {
	var zeroHashes = [][numHashBytes]byte{
		initialZeroHash,
	}
	// This generates a leaf = HashZero in position 0. In the rest of the positions that are equivalent to the ascending levels,
	// we set the hashes of the nodes. So all nodes from level i=5 will have the same value and same children nodes.
	for i := 1; i <= int(height); i++ {
		zeroHashes = append(zeroHashes, Hash(zeroHashes[i-1], zeroHashes[i-1]))
	}
	return zeroHashes
}

type KeccakMerkleProof struct {
	Siblings [][numHashBytes]byte
}

func (proof *KeccakMerkleProof) GetRoot() []byte {
	var buf []byte
	for _, sibling := range proof.Siblings {
		buf = append(buf, sibling[:]...)
	}
	return buf
}

func (proof *KeccakMerkleProof) GetMerkleRoot(
	index int,
	leafHash common.Hash,
) common.Hash {
	nodeHash := [numHashBytes]byte{}
	copy(nodeHash[:], leafHash[:])

	for _, sibling := range proof.Siblings {
		if index&1 == 1 {
			nodeHash = Hash(sibling, nodeHash)
		} else {
			nodeHash = Hash(nodeHash, sibling)
		}
		index >>= 1
	}

	return nodeHash
}

func (proof *KeccakMerkleProof) Verify(
	root common.Hash,
	index int,
	leafHash common.Hash,
) error {
	expectedRoot := proof.GetMerkleRoot(index, leafHash)

	if expectedRoot != root {
		var ErrMerkleProofVerifyFail = errors.New("the Merkle proof verify fail")
		return ErrMerkleProofVerifyFail
	}

	return nil
}
