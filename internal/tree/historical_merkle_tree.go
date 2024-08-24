package tree

import (
	"encoding/json"
	"fmt"
	"math/bits"
)

type PoseidonSubTreePreimage struct {
	Height             uint8
	NonDefaultChildren []string
	DefaultChild       string
}

const (
	preimageTypeEmpty = iota
	preimageTypePoseidonSubTree
	preimageTypeKeccakSubTree
)

type PreimageWithType struct {
	Type int
	// JSON
	PreimageJSON []byte
}

type Preimages interface {
	BatchUpdate(preimages map[string]*PreimageWithType) error
	GetPreimageByHash(hash string) (*PreimageWithType, error)
}

type preimagesOnMemory map[string]*PreimageWithType

func (p *preimagesOnMemory) BatchUpdate(preimages map[string]*PreimageWithType) error {
	for k, v := range preimages {
		(*p)[k] = v
	}

	return nil
}

func (p *preimagesOnMemory) GetPreimageByHash(k string) (*PreimageWithType, error) {
	v, ok := (*p)[k]
	if !ok {
		return nil, fmt.Errorf("preimage not found")
	}

	return v, nil
}

type HistoricalPoseidonMerkleTree struct {
	*PoseidonMerkleTree
	NextUnusedIndex      int
	HistoricalRoots      []string
	Cache                Preimages
	CachingSubTreeHeight uint8
}

func NewHistoricalPoseidonMerkleTree(
	height uint8,
	zeroHash *PoseidonHashOut,
	cachingSubTreeHeight uint8,
) (*HistoricalPoseidonMerkleTree, error) {
	mt, err := NewPoseidonMerkleTree(height, zeroHash)
	if err != nil {
		return nil, err
	}

	root := mt.GetRoot()
	cache := make(preimagesOnMemory)
	addCache(
		root.String(),
		height,
		[]string{},
		zeroHash.String(),
		cache,
	)

	return &HistoricalPoseidonMerkleTree{
		PoseidonMerkleTree:   mt,
		NextUnusedIndex:      0,
		HistoricalRoots:      []string{root.String()},
		Cache:                &cache,
		CachingSubTreeHeight: cachingSubTreeHeight,
	}, nil
}

type PoseidonMerkleLeafWithIndex struct {
	Index    int
	LeafHash *PoseidonHashOut
}

func (t *HistoricalPoseidonMerkleTree) UpdateLeaves(
	leaves []*PoseidonMerkleLeafWithIndex,
) (root *PoseidonHashOut, err error) {
	tmpPreimages := make(preimagesOnMemory)
	for _, leaf := range leaves {

		_, err := t.updateLeaf(leaf.Index, leaf.LeafHash, tmpPreimages)
		if err != nil {
			return nil, err
		}
	}

	t.Cache.BatchUpdate(tmpPreimages)

	root = t.PoseidonMerkleTree.GetRoot()
	t.HistoricalRoots = append(t.HistoricalRoots, root.String())

	return root, nil
}

func (t *HistoricalPoseidonMerkleTree) updateLeaf(
	index int,
	leafHash *PoseidonHashOut,
	cache preimagesOnMemory,
) (string, error) {
	t.PoseidonMerkleTree.updateLeaf(index, new(PoseidonHashOut).Set(leafHash))

	if int(index) >= t.NextUnusedIndex {
		t.NextUnusedIndex = int(index) + 1
	}

	significantHeight := uint8(effectiveBits(uint(t.NextUnusedIndex - 1)))

	restHeight := significantHeight % t.CachingSubTreeHeight

	numTargetNodes := 1 << t.CachingSubTreeHeight
	clearMask := numTargetNodes - 1

	targetChildNodeIndex := 1<<t.height + index
	subTreeHeight := t.height
	for i := uint8(0); i < significantHeight-restHeight; i += t.CachingSubTreeHeight {
		targetLeftMostChildNodeIndex := targetChildNodeIndex & ^clearMask
		nonDefaultChildren := []string{}
		for j := 0; j < numTargetNodes; j++ {
			nodeHash := t.GetNodeHash(targetLeftMostChildNodeIndex + j)
			nonDefaultChildren = append(nonDefaultChildren, nodeHash.String())
		}

		nextTargetChildNodeIndex := targetChildNodeIndex >> t.CachingSubTreeHeight
		parentHash := t.GetNodeHash(nextTargetChildNodeIndex)
		addCache(
			parentHash.String(),
			t.CachingSubTreeHeight,
			nonDefaultChildren,
			t.getZeroHash(targetChildNodeIndex).String(),
			cache,
		)
		targetChildNodeIndex = nextTargetChildNodeIndex
		subTreeHeight -= t.CachingSubTreeHeight
	}

	numTargetNodes = 1 << restHeight
	clearMask = numTargetNodes - 1

	targetLeftMostChildNodeIndex := targetChildNodeIndex & ^clearMask
	nonDefaultChildren := []string{}
	for j := 0; j < numTargetNodes; j++ {
		nodeHash := t.GetNodeHash(targetLeftMostChildNodeIndex + j)
		nonDefaultChildren = append(nonDefaultChildren, nodeHash.String())
	}

	nextTargetParentNodeIndex := targetChildNodeIndex >> subTreeHeight
	if nextTargetParentNodeIndex != 1 {
		// Fatal error
		return "", fmt.Errorf("node index of root hash must be 1, but got %d", nextTargetParentNodeIndex)
	}

	root := t.PoseidonMerkleTree.GetRoot()
	addCache(
		root.String(),
		subTreeHeight,
		nonDefaultChildren,
		t.getZeroHash(targetChildNodeIndex).String(),
		cache,
	)

	return root.String(), nil
}

func (t *HistoricalPoseidonMerkleTree) Prove(targetRoot *PoseidonHashOut, index int) (*PoseidonMerkleProof, error) {
	nodeIndex := 1<<t.height + index

	siblings := make([]*PoseidonHashOut, 0)

	targetHeight := uint8(0)
	targetNodeHash := new(PoseidonHashOut).Set(targetRoot)
	for targetHeight < t.height {
		preimageWithType, err := t.Cache.GetPreimageByHash(targetNodeHash.String())
		if err != nil {
			return nil, fmt.Errorf("preimage not found")
		}

		if preimageWithType.Type != preimageTypePoseidonSubTree {
			return nil, fmt.Errorf("unsupported preimage type")
		}

		var preimage PoseidonSubTreePreimage
		err = json.Unmarshal(preimageWithType.PreimageJSON, &preimage)
		if err != nil {
			return nil, err
		}

		defaultChild := new(PoseidonHashOut)
		err = defaultChild.FromString(preimage.DefaultChild)
		if err != nil {
			return nil, err
		}

		leaves := []*PoseidonHashOut{}
		for _, leaf := range preimage.NonDefaultChildren {
			leafHash := new(PoseidonHashOut)
			err = leafHash.FromString(leaf)
			if err != nil {
				return nil, err
			}

			leaves = append(leaves, leafHash)
		}

		mt, err := NewPoseidonIncrementalMerkleTree(uint8(preimage.Height), leaves, defaultChild)
		if err != nil {
			return nil, err
		}

		// remove `height` bits from the left without lest-most bits
		subTreeIndex := getSubTreeIndex(nodeIndex, targetHeight, preimage.Height)
		newSiblings, _, err := mt.ComputeMerkleProof(uint64(subTreeIndex), leaves)
		if err != nil {
			return nil, err
		}

		siblings = append(newSiblings, siblings...)
		targetHeight += preimage.Height
		if subTreeIndex >= len(leaves) {
			targetNodeHash = defaultChild
		} else {
			targetNodeHash = leaves[subTreeIndex]
		}
	}

	if targetHeight != t.height {
		return nil, fmt.Errorf("target height must be equal to tree height %d, but got %d", t.height, targetHeight)
	}

	return &PoseidonMerkleProof{
		Siblings: siblings,
	}, nil
}

// getSubTreeIndex calculates the index of a subtree within a larger tree structure
//
// root              <--------------------------height------------------------------>   leaf
// root              <--truncatedHeight---><---------------maskWidth---------------->   leaf
// root              <--truncatedHeight---><----subTreeHeight----><-remainingHeight->   leaf
// nodeIndex    = 0b10010000111010101011101010001010000011100110010100101101010011010
// mask         =                        0b111111111111111111111110000000000000000000
// subTreeIndex =                        0b01000101000001110011001
func getSubTreeIndex(nodeIndex int, truncatedHeight, subTreeHeight uint8) int {
	heightInt := bits.Len(uint(nodeIndex)) - 1
	const int8Key = 8
	if heightInt >= (1 << int8Key) {
		panic("height must be less than 256")
	}

	height := uint8(heightInt)
	if truncatedHeight+subTreeHeight > height {
		panic("truncatedHeight + subTreeHeight must be less than or equal to height")
	}

	remainingHeight := height - truncatedHeight - subTreeHeight
	maskWidth := height - truncatedHeight
	mask := 1<<maskWidth - 1<<remainingHeight
	subTreeIndex := nodeIndex & mask
	subTreeIndex >>= remainingHeight

	return subTreeIndex
}

func addCache(
	root string,
	treeHeight uint8,
	nonDefaultChildren []string,
	defaultChild string,
	cache preimagesOnMemory,
) error {
	subTreePreimage := PoseidonSubTreePreimage{
		Height:             treeHeight,
		NonDefaultChildren: nonDefaultChildren,
		DefaultChild:       defaultChild,
	}

	subTreePreimageJSON, err := json.Marshal(subTreePreimage)
	if err != nil {
		return err
	}

	subTree := PreimageWithType{
		Type:         preimageTypePoseidonSubTree,
		PreimageJSON: subTreePreimageJSON,
	}

	cache[root] = &subTree

	return nil
}

// log2Ceil
func effectiveBits(n uint) uint32 {
	if n == 0 {
		return 0
	}

	bits := uint32(0)
	for n > 0 {
		n >>= 1
		bits++
	}

	return bits
}
