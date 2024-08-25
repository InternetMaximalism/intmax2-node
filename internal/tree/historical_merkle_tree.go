package tree

import (
	"encoding/json"
	"fmt"
	"intmax2-node/internal/hash/goldenposeidon"
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

type MerkleTreeHistory interface {
	PushVersion(root *PoseidonHashOut, leaves map[int]*PoseidonHashOut, preimages map[string]*PreimageWithType) error
	LatestVersion() int
	GetPreimageByHash(hash string) (*PreimageWithType, error)
	GetLeaves(version int, indices []int) (map[int]*PoseidonHashOut, error)
	GetRoot(version int) (*PoseidonHashOut, error)
	SetUnusedIndex(index int)
	NextUnusedIndex() int
}

type merkleTreeHistoryOnMemory struct {
	preimages       map[string]*PreimageWithType
	versionedLeaves []map[int]*PoseidonHashOut
	historicalRoots []string
	nextUnusedIndex int
	countWrite      int
	countRead       int
}

func NewMerkleTreeHistoryOnMemory() *merkleTreeHistoryOnMemory {
	return &merkleTreeHistoryOnMemory{
		preimages:       make(map[string]*PreimageWithType),
		versionedLeaves: make([]map[int]*PoseidonHashOut, 0), // version => leaf => hash
	}
}

func (p *merkleTreeHistoryOnMemory) PushVersion(root *PoseidonHashOut, leaves map[int]*PoseidonHashOut, preimages map[string]*PreimageWithType) error {
	for k, v := range preimages {
		p.preimages[k] = v
	}

	p.versionedLeaves = append(p.versionedLeaves, leaves)
	p.historicalRoots = append(p.historicalRoots, root.String())
	p.countWrite += len(preimages)

	return nil
}

func (p *merkleTreeHistoryOnMemory) LatestVersion() int {
	return len(p.versionedLeaves) - 1
}

func (p *merkleTreeHistoryOnMemory) GetPreimageByHash(k string) (*PreimageWithType, error) {
	v, ok := p.preimages[k]
	if !ok {
		return nil, fmt.Errorf("preimage not found")
	}
	p.countRead++

	return v, nil
}

func (p *merkleTreeHistoryOnMemory) GetLeaves(version int, indices []int) (map[int]*PoseidonHashOut, error) {
	if version < 0 {
		return nil, fmt.Errorf("version must be greater than or equal to 0")
	}
	if version >= len(p.versionedLeaves) {
		return nil, fmt.Errorf("version not found")
	}

	result := make(map[int]*PoseidonHashOut)
	for _, index := range indices {
		var hash *PoseidonHashOut
		var v int
		for v := version; v >= 0; v-- {
			leaves := p.versionedLeaves[v]
			var ok bool
			foundHash, ok := leaves[index]
			if ok {
				hash = new(PoseidonHashOut).Set(foundHash)
				break
			}
		}

		if v < 0 {
			continue
		}
		p.countRead++

		result[index] = hash
	}

	return result, nil
}

func (p *merkleTreeHistoryOnMemory) GetRoot(version int) (*PoseidonHashOut, error) {
	if version < 0 {
		return nil, fmt.Errorf("version must be greater than or equal to 0")
	}
	if version >= len(p.historicalRoots) {
		return nil, fmt.Errorf("version not found")
	}

	root := p.historicalRoots[version]
	if root == "" {
		return nil, fmt.Errorf("root not found")
	}

	p.countRead++

	result := new(PoseidonHashOut)
	result.FromString(root)

	return result, nil
}

func (p *merkleTreeHistoryOnMemory) SetUnusedIndex(index int) {
	p.nextUnusedIndex = index
}

func (p *merkleTreeHistoryOnMemory) NextUnusedIndex() int {
	return p.nextUnusedIndex
}

func (p *merkleTreeHistoryOnMemory) Size() int {
	size := 0
	for key, value := range p.preimages {
		size += len(key) + len(value.PreimageJSON) + 16
	}

	return size
}

func (p *merkleTreeHistoryOnMemory) ReportStats() {
	fmt.Printf("size of Storage: %d bytes\n", p.Size())
	fmt.Printf("count of writing storage: %d\n", p.countRead)
	fmt.Printf("count of reading storage: %d\n", p.countRead)
}

type HistoricalPoseidonMerkleTree struct {
	*PoseidonMerkleTree
	Storage              MerkleTreeHistory
	CachingSubTreeHeight uint8
}

func NewHistoricalPoseidonMerkleTree(
	height uint8,
	zeroHash *PoseidonHashOut,
	storage MerkleTreeHistory,
	cachingSubTreeHeight uint8,
) (*HistoricalPoseidonMerkleTree, error) {
	mt, err := NewPoseidonMerkleTree(height, zeroHash)
	if err != nil {
		return nil, err
	}

	root := mt.GetRoot()
	tmpCache := make(map[string]*PreimageWithType)
	addCache(
		root.String(),
		height,
		[]string{},
		zeroHash.String(),
		tmpCache,
	)

	emptyLeaves := make(map[int]*goldenposeidon.PoseidonHashOut)
	storage.PushVersion(root, emptyLeaves, tmpCache)

	tree := HistoricalPoseidonMerkleTree{
		PoseidonMerkleTree:   mt,
		Storage:              storage,
		CachingSubTreeHeight: cachingSubTreeHeight,
	}

	nextUnusedIndex := storage.NextUnusedIndex()
	indices := []int{}
	for i := 0; i < nextUnusedIndex; i++ {
		indices = append(indices, i)
	}

	err = tree.LoadNodeHashes(indices)
	if err != nil {
		return nil, err
	}

	return &tree, nil
}

type PoseidonMerkleLeafWithIndex struct {
	Index    int
	LeafHash *PoseidonHashOut
}

func (t *HistoricalPoseidonMerkleTree) UpdateLeaves(
	leaves []*PoseidonMerkleLeafWithIndex,
) (root *PoseidonHashOut, err error) {
	tmpPreimages := make(map[string]*PreimageWithType)
	for _, leaf := range leaves {
		_, err := t.updateLeaf(leaf.Index, leaf.LeafHash, tmpPreimages)
		if err != nil {
			return nil, err
		}
	}

	leavesMap := make(map[int]*PoseidonHashOut)
	for _, leaf := range leaves {
		leavesMap[leaf.Index] = leaf.LeafHash
	}

	t.Storage.PushVersion(t.PoseidonMerkleTree.GetRoot(), leavesMap, tmpPreimages)

	root = t.PoseidonMerkleTree.GetRoot()
	// t.Storage.historicalRoots = append(t.Storage.historicalRoots, root.String())

	return root, nil
}

func (t *HistoricalPoseidonMerkleTree) GetRoot() *PoseidonHashOut {
	return t.PoseidonMerkleTree.GetRoot()
}

func (t *HistoricalPoseidonMerkleTree) LoadNodeHashes(indices []int) error {
	latestVersion := t.Storage.LatestVersion()
	root, err := t.Storage.GetRoot(latestVersion)
	if err != nil {
		fmt.Printf("not found root root")
		return err
	}

	leavesMap, err := t.Storage.GetLeaves(latestVersion, indices)
	if err != nil {
		return err
	}

	leaves := make([]*PoseidonMerkleLeafWithIndex, 0, len(leavesMap))
	// for i := range indices {
	// 	leaves[i] = &PoseidonMerkleLeafWithIndex{
	// 		Index:    indices[i],
	// 		LeafHash: leavesMap[indices[i]],
	// 	}
	// }
	for k, v := range leavesMap {
		v := PoseidonMerkleLeafWithIndex{
			Index:    k,
			LeafHash: v,
		}
		leaves = append(leaves, &v)
	}

	t.PoseidonMerkleTree.ClearCache()
	_, err = t.UpdateLeaves(leaves)
	if err != nil {
		return err
	}

	if t.PoseidonMerkleTree.GetRoot().Equal(root) {
		return nil
	}

	return nil
}

func (t *HistoricalPoseidonMerkleTree) updateLeaf(
	index int,
	leafHash *PoseidonHashOut,
	cache map[string]*PreimageWithType,
) (string, error) {
	// t.LoadNodeHashes()

	t.PoseidonMerkleTree.updateLeaf(index, new(PoseidonHashOut).Set(leafHash))

	nextUnusedIndex := t.Storage.NextUnusedIndex()
	if int(index) >= nextUnusedIndex {
		nextUnusedIndex = int(index) + 1
		t.Storage.SetUnusedIndex(int(index) + 1)
	}

	significantHeight := uint8(effectiveBits(uint(nextUnusedIndex - 1)))

	restHeight := significantHeight % t.CachingSubTreeHeight

	numTargetNodes := 1 << t.CachingSubTreeHeight
	clearMask := numTargetNodes - 1

	targetChildNodeIndex := 1<<t.height + index
	subTreeHeight := t.height
	for i := uint8(0); i < significantHeight-restHeight; i += t.CachingSubTreeHeight {
		targetLeftMostChildNodeIndex := targetChildNodeIndex & ^clearMask
		nonDefaultChildren := []string{}
		defaultChild := t.getZeroHash(targetChildNodeIndex)
		for j := 0; j < numTargetNodes; j++ {
			if nodeHash, ok := t.nodeHashes[targetLeftMostChildNodeIndex+j]; ok {
				nonDefaultChildren = append(nonDefaultChildren, nodeHash.String())
			}
		}

		nextTargetChildNodeIndex := targetChildNodeIndex >> t.CachingSubTreeHeight
		parentHash := t.GetNodeHash(nextTargetChildNodeIndex)
		addCache(
			parentHash.String(),
			t.CachingSubTreeHeight,
			nonDefaultChildren,
			defaultChild.String(),
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
	// t.LoadNodeHashes()

	nodeIndex := 1<<t.height + index

	siblings := make([]*PoseidonHashOut, 0)

	targetHeight := uint8(0)
	targetNodeHash := new(PoseidonHashOut).Set(targetRoot)
	for targetHeight < t.height {
		preimageWithType, err := t.Storage.GetPreimageByHash(targetNodeHash.String())
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
		// fmt.Printf("preimage = %v\n", preimage)

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
	parentHash string,
	treeHeight uint8,
	nonDefaultChildren []string,
	defaultChild string,
	cache map[string]*PreimageWithType,
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

	cache[parentHash] = &subTree

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
