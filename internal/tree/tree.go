package tree

import (
	"fmt"
	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/ethereum/go-ethereum/log"
)

// type KeccakMerkleTree struct {
// 	height      uint8
// 	zeroHashes  []common.Hash
// 	count       uint32
// 	siblings    []common.Hash
// 	currentRoot common.Hash
// }

type poseidonHashOut = goldenposeidon.PoseidonHashOut

type PoseidonMerkleTree struct {
	height      uint8
	zeroHashes  []*poseidonHashOut
	count       uint64
	siblings    []*poseidonHashOut
	currentRoot poseidonHashOut
}

// NewPoseidonMerkleTreeWithLeaves creates new PoseidonMerkleTree by giving leaf nodes.
func NewPoseidonMerkleTree(height uint8, initialLeaves []*poseidonHashOut, zeroHash *poseidonHashOut) (*PoseidonMerkleTree, error) {
	mt := &PoseidonMerkleTree{
		zeroHashes: generateZeroHashes(height, zeroHash),
		height:     height,
		count:      uint64(len(initialLeaves)),
	}
	var err error
	mt.siblings, mt.currentRoot, err = mt.initSiblings(initialLeaves)
	if err != nil {
		log.Error("error initializing si siblings. Error: ", err)
		return nil, err
	}
	log.Debug("Initial count: ", mt.count)
	log.Debug("Initial root: ", mt.currentRoot)
	return mt, nil
}

func buildIntermediate(leaves []*poseidonHashOut) (nodes [][]*poseidonHashOut, hashes []*poseidonHashOut) {
	// var (
	// 	nodes  [][]*poseidonHashOut
	// 	hashes []*poseidonHashOut
	// )
	for i := 0; i < len(leaves); i += 2 {
		var left, right = i, i + 1
		h := goldenposeidon.Compress(leaves[left], leaves[right])
		nodes = append(nodes, []*poseidonHashOut{h, leaves[left], leaves[right]})
		hashes = append(hashes, h)
	}
	return nodes, hashes
}

// BuildMerkleRoot computes the root given the leaves of the tree
func (mt *PoseidonMerkleTree) BuildMerkleRoot(leaves []*poseidonHashOut) (*poseidonHashOut, error) {
	var (
		nodes [][][]*poseidonHashOut
		ns    [][]*poseidonHashOut
	)
	if len(leaves) == 0 {
		leaves = append(leaves, mt.zeroHashes[0])
	}

	for h := uint8(0); h < mt.height; h++ {
		if len(leaves)%2 == 1 {
			leaves = append(leaves, mt.zeroHashes[h])
		}
		ns, leaves = buildIntermediate(leaves)
		nodes = append(nodes, ns)
	}
	if len(ns) != 1 {
		return nil, fmt.Errorf("error: more than one root detected: %+v", nodes)
	}

	return ns[0][0], nil
}

func generateZeroHashes(height uint8, zeroHash *poseidonHashOut) []*poseidonHashOut {
	var zeroHashes = []*poseidonHashOut{
		new(poseidonHashOut).Set(zeroHash),
	}
	// This generates a leaf = HashZero in position 0. In the rest of the positions that are equivalent to the ascending levels,
	// we set the hashes of the nodes. So all nodes from level i=5 will have the same value and same children nodes.
	for i := 1; i <= int(height); i++ {
		zeroHashes = append(zeroHashes, goldenposeidon.Compress(zeroHashes[i-1], zeroHashes[i-1]))
	}
	return zeroHashes
}

// ComputeMerkleProof computes the merkleProof and root given the leaves of the tree
func (mt *PoseidonMerkleTree) ComputeMerkleProof(index uint64, leaves []*poseidonHashOut) (siblings []*poseidonHashOut, root poseidonHashOut, err error) {
	var ns [][]*poseidonHashOut
	if len(leaves) == 0 {
		leaves = append(leaves, mt.zeroHashes[0])
	}
	proofIndex := index
	for h := uint8(0); h < mt.height; h++ {
		if len(leaves)%2 == 1 {
			leaves = append(leaves, mt.zeroHashes[h])
		}
		if proofIndex%2 == 1 {
			// If it is odd
			siblings = append(siblings, leaves[proofIndex-1])
		} else if len(leaves) > 1 {
			if proofIndex >= uint64(len(leaves)) {
				siblings = append(siblings, leaves[proofIndex-1])
			} else {
				siblings = append(siblings, leaves[proofIndex+1])
			}
		}

		var (
			nsi    [][]*poseidonHashOut
			hashes []*poseidonHashOut
		)
		for i := 0; i < len(leaves); i += 2 {
			var left, right = i, i + 1
			h := goldenposeidon.Compress(leaves[left], leaves[right])
			nsi = append(nsi, []*poseidonHashOut{h, leaves[left], leaves[right]})
			hashes = append(hashes, h)
		}
		// Find the index of the leave in the next level of the tree.
		// Divide the index by 2 to find the position in the upper level
		const half = 2
		proofIndex = uint64(float64(proofIndex) / half)
		ns = nsi
		leaves = hashes
	}
	if len(ns) != 1 {
		return nil, poseidonHashOut{}, fmt.Errorf("error: more than one root detected: %+v", ns)
	}

	return siblings, *ns[0][0], nil
}

// AddLeaf adds new leaves to the tree and computes the new root
func (mt *PoseidonMerkleTree) AddLeaf(index uint64, leaf *poseidonHashOut) (*poseidonHashOut, error) {
	if index != mt.count {
		return nil, fmt.Errorf("mismatched leaf count: %d, expected: %d", index, mt.count)
	}
	cur := new(poseidonHashOut).Set(leaf)
	isFilledSubTree := true

	for h := uint8(0); h < mt.height; h++ {
		if index&(1<<h) > 0 {
			child := cur
			parent := goldenposeidon.Compress(mt.siblings[h], child)
			cur = parent
		} else {
			if isFilledSubTree {
				// we will update the sibling when the sub tree is complete
				mt.siblings[h] = new(poseidonHashOut).Set(cur)
				// we have a left child in this layer, it means the right child is empty so the sub tree is not completed
				isFilledSubTree = false
			}
			child := cur
			parent := goldenposeidon.Compress(child, mt.zeroHashes[h])
			cur = parent
			// the sibling of 0 bit should be the zero hash, since we are in the last node of the tree
		}
	}
	mt.currentRoot.Set(cur)
	mt.count++
	return cur, nil
}

// initSiblings returns the siblings of the node at the given index.
// it is used to initialize the siblings array in the beginning.
func (mt *PoseidonMerkleTree) initSiblings(initialLeaves []*poseidonHashOut) (siblings []*poseidonHashOut, root poseidonHashOut, err error) {
	if mt.count != uint64(len(initialLeaves)) {
		return nil, poseidonHashOut{}, fmt.Errorf("error: mt.count and initialLeaves length mismatch")
	}
	if mt.count == 0 {
		for h := 0; h < int(mt.height); h++ {
			left := new(poseidonHashOut)
			copy(left.Elements[:], mt.zeroHashes[h].Elements[:])
			siblings = append(siblings, left)
		}
		root, err := mt.BuildMerkleRoot(initialLeaves) // nolint:govet
		if err != nil {
			log.Error("error calculating initial root: ", err)
			return nil, poseidonHashOut{}, err
		}
		return siblings, *root, nil
	}

	return mt.ComputeMerkleProof(mt.count, initialLeaves)
}

// GetCurrentRootCountAndSiblings returns the latest root, count and sibblings
func (mt *PoseidonMerkleTree) GetCurrentRootCountAndSiblings() (root poseidonHashOut, count uint64, siblings []*poseidonHashOut) {
	return mt.currentRoot, mt.count, mt.siblings
}
