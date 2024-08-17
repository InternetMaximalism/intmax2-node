package tree

import (
	"errors"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type LeafIndex = int

type IndexedMerkleLeaf struct {
	Key       *big.Int  `json:"key"`
	Value     uint64    `json:"value"`
	NextIndex LeafIndex `json:"nextIndex"`
	NextKey   *big.Int  `json:"nextKey"`
}

func (leaf *IndexedMerkleLeaf) Set(other *IndexedMerkleLeaf) *IndexedMerkleLeaf {
	leaf.NextIndex = other.NextIndex
	leaf.Key = new(big.Int).Set(other.Key)
	leaf.NextKey = new(big.Int).Set(other.NextKey)
	leaf.Value = other.Value

	return leaf
}

func (leaf *IndexedMerkleLeaf) SetDefault() *IndexedMerkleLeaf {
	leaf.NextIndex = 0
	leaf.Key = new(big.Int)
	leaf.NextKey = new(big.Int)
	leaf.Value = 0

	return leaf
}

func (leaf *IndexedMerkleLeaf) EmptyLeaf() *IndexedMerkleLeaf {
	leaf.NextIndex = 0
	leaf.Key = new(big.Int)
	leaf.NextKey = new(big.Int)
	leaf.Value = 0

	return leaf
}

func (leaf *IndexedMerkleLeaf) ToFieldElementSlice() []ffg.Element {
	res := finite_field.NewBuffer(make([]ffg.Element, 0))
	err := finite_field.WriteUint64(res, uint64(leaf.NextIndex))
	if err != nil {
		panic(err)
	}
	key := new(intMaxTypes.Uint256).FromBigInt(leaf.Key).ToFieldElementSlice()
	for i := range key {
		finite_field.WriteGoldilocksField(res, &key[i])
	}
	nextKey := new(intMaxTypes.Uint256).FromBigInt(leaf.NextKey).ToFieldElementSlice()
	for i := range nextKey {
		finite_field.WriteGoldilocksField(res, &nextKey[i])
	}
	err = finite_field.WriteUint64(res, leaf.Value)
	if err != nil {
		panic(err)
	}

	return res.Inner()
}

func (leaf *IndexedMerkleLeaf) Hash() *goldenposeidon.PoseidonHashOut {
	return goldenposeidon.HashNoPad(leaf.ToFieldElementSlice())
}

type IndexedMerkleProof struct {
	Siblings []*goldenposeidon.PoseidonHashOut `json:"siblings"`
}

// TODO: leaf is *BlockHashLeaf?
func (proof *IndexedMerkleProof) GetRoot(leaf *IndexedMerkleLeaf, index LeafIndex) *goldenposeidon.PoseidonHashOut {
	height := len(proof.Siblings)
	if index >= 1<<height {
		panic("index out of bounds")
	}
	nodeIndex := 1<<height + index
	h := new(PoseidonHashOut).Set(leaf.Hash())

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

// TODO: leaf is *BlockHashLeaf?
func (proof *IndexedMerkleProof) Verify(leaf *IndexedMerkleLeaf, index LeafIndex, root *goldenposeidon.PoseidonHashOut) error {
	computedRoot := proof.GetRoot(leaf, index)
	if !computedRoot.Equal(root) {
		return errors.New("invalid root")
	}

	return nil
}

type IndexedInsertionProof struct {
	Index        LeafIndex
	LowLeafProof IndexedMerkleProof
	LeafProof    IndexedMerkleProof
	LowLeafIndex LeafIndex
	PrevLowLeaf  IndexedMerkleLeaf
}

func (proof *IndexedInsertionProof) GetNewRoot(
	key *big.Int,
	value uint64,
	prevRoot *PoseidonHashOut,
) (*PoseidonHashOut, error) {
	if proof.PrevLowLeaf.Key.Cmp(key) == 1 {
		return nil, errors.New("key is not lower-bounded")
	}

	if key.Cmp(proof.PrevLowLeaf.NextKey) == 1 || proof.PrevLowLeaf.NextKey.Cmp(new(big.Int)) == 0 {
		return nil, errors.New("key is not upper-bounded")
	}

	err := proof.LowLeafProof.Verify(&proof.PrevLowLeaf, proof.LowLeafIndex, prevRoot)
	if err != nil {
		return nil, err
	}

	newLowLeaf := IndexedMerkleLeaf{
		NextIndex: proof.Index,
		NextKey:   key,
		Key:       proof.PrevLowLeaf.Key,
		Value:     proof.PrevLowLeaf.Value,
	}
	tempRoot := proof.LowLeafProof.GetRoot(&newLowLeaf, proof.LowLeafIndex)
	err = proof.LeafProof.Verify(
		new(IndexedMerkleLeaf).EmptyLeaf(),
		proof.Index,
		tempRoot,
	)
	if err != nil {
		return nil, err
	}

	leaf := IndexedMerkleLeaf{
		NextIndex: proof.PrevLowLeaf.NextIndex,
		Key:       key,
		NextKey:   proof.PrevLowLeaf.NextKey,
		Value:     value,
	}

	return proof.LeafProof.GetRoot(&leaf, proof.Index), nil
}

func (proof *IndexedInsertionProof) ConditionalGetNewRoot(condition bool, key *big.Int, value uint64, prevRoot *PoseidonHashOut) (*PoseidonHashOut, error) {
	if condition {
		return proof.GetNewRoot(key, value, prevRoot)
	}

	return prevRoot, nil
}

type IndexedUpdateProof struct {
	LeafProof IndexedMerkleProof
	LeafIndex LeafIndex
	PrevLeaf  IndexedMerkleLeaf
}

func (proof *IndexedUpdateProof) GetNewRoot(key *big.Int, prevValue, newValue uint64, prevRoot *PoseidonHashOut) (*PoseidonHashOut, error) {
	if proof.PrevLeaf.Value != prevValue {
		return nil, errors.New("value mismatch")
	}

	if proof.PrevLeaf.Key.Cmp(key) != 0 {
		return nil, errors.New("key mismatch")
	}

	err := proof.LeafProof.Verify(&proof.PrevLeaf, proof.LeafIndex, prevRoot)
	if err != nil {
		return nil, err
	}

	newLeaf := IndexedMerkleLeaf{
		Value:     newValue,
		NextIndex: proof.PrevLeaf.NextIndex,
		Key:       proof.PrevLeaf.Key,
		NextKey:   proof.PrevLeaf.NextKey,
	}
	return proof.LeafProof.GetRoot(&newLeaf, proof.LeafIndex), nil
}

func (proof *IndexedUpdateProof) Verify(key *big.Int, prevValue, newValue uint64, prevRoot, newRoot *PoseidonHashOut) error {
	expectedNewRoot, err := proof.GetNewRoot(key, prevValue, newValue, prevRoot)
	if err != nil {
		return err
	}

	if !newRoot.Equal(expectedNewRoot) {
		return errors.New("new_root mismatch")
	}

	return nil
}

type IndexedMerkleTree struct {
	Leaves []*IndexedMerkleLeaf
	inner  *PoseidonMerkleTree
}

type IndexedMembershipProof struct {
	IsIncluded bool               `json:"isIncluded"`
	LeafProof  IndexedMerkleProof `json:"leafProof"`
	LeafIndex  LeafIndex          `json:"leafIndex"`
	Leaf       IndexedMerkleLeaf  `json:"leaf"`
}

func NewIndexedMerkleTree(height uint8, zeroHash *goldenposeidon.PoseidonHashOut) (*IndexedMerkleTree, error) {
	tree, err := NewPoseidonMerkleTree(height, nil, zeroHash)
	if err != nil {
		return nil, err
	}

	defaultLeaf := new(IndexedMerkleLeaf).SetDefault()
	defaultLeafHash := defaultLeaf.Hash()
	tree.updateLeaf(0, defaultLeafHash)

	return &IndexedMerkleTree{
		Leaves: []*IndexedMerkleLeaf{defaultLeaf},
		inner:  tree,
	}, nil
}

func (t *IndexedMerkleTree) GetRoot() PoseidonHashOut {
	root := t.inner.GetRoot()

	return *root
}

func (t *IndexedMerkleTree) GetLeaf(index LeafIndex) *IndexedMerkleLeaf {
	return t.Leaves[index]
}

func (t *IndexedMerkleTree) Prove(index LeafIndex) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	siblings, err = t.inner.Prove(index)
	if err != nil {
		return nil, goldenposeidon.PoseidonHashOut{}, err
	}

	root = t.GetRoot()

	return siblings, root, err
}

func (t *IndexedMerkleTree) ProveMembership(key *big.Int) (membership_proof *IndexedMembershipProof, root PoseidonHashOut, err error) {
	lowIndex := t.GetLowIndex(key)
	lowLeaf := t.GetLeaf(lowIndex)
	leafProof, root, err := t.Prove(lowIndex)
	if err != nil {
		return nil, goldenposeidon.PoseidonHashOut{}, err
	}

	membership_proof = &IndexedMembershipProof{
		IsIncluded: lowLeaf.Key.Cmp(key) == 0,
		LeafProof:  IndexedMerkleProof{Siblings: leafProof},
		LeafIndex:  lowIndex,
		Leaf:       *lowLeaf,
	}

	return membership_proof, root, nil
}

func (t *IndexedMerkleTree) GetLowIndex(key *big.Int) LeafIndex {
	validLeafCandidates := make([]LeafIndex, 0)
	for i, leaf := range t.Leaves {
		// key >= leaf.Key
		isValidLowerLimit := key.Cmp(leaf.Key) != -1
		// key < leaf.NextKey || leaf.NextKey == 0
		isValidUpperLimit := key.Cmp(leaf.NextKey) == -1 || leaf.NextKey.Cmp(big.NewInt(0)) == 0
		if isValidLowerLimit && isValidUpperLimit {
			validLeafCandidates = append(validLeafCandidates, i)
		}
	}

	if len(validLeafCandidates) == 0 {
		panic("validLeafCandidates should not be zero")
	}
	if len(validLeafCandidates) > 1 {
		panic("too many candidates")
	}

	return validLeafCandidates[0]
}

func (t *IndexedMerkleTree) GetIndex(key *big.Int) (LeafIndex, bool) {
	for i, leaf := range t.Leaves {
		if key.Cmp(leaf.Key) == 0 {
			return LeafIndex(i), true // nolint:unconvert
		}
	}

	return 0, false
}

func (t *IndexedMerkleTree) Update(key *big.Int, value uint64) (*IndexedUpdateProof, error) {
	if key.Cmp(big.NewInt(0)) == -1 {
		return nil, errors.New("key is negative")
	}
	const uint256Key uint = 256
	if key.Cmp(new(big.Int).Lsh(big.NewInt(1), uint256Key)) == 1 {
		return nil, errors.New("key is too large")
	}

	index, ok := t.GetIndex(key)
	if !ok {
		return nil, errors.New("key doesn't exist")
	}

	prevLeaf := new(IndexedMerkleLeaf).Set(t.GetLeaf(index))

	t.Leaves[index].Value = value
	t.inner.updateLeaf(index, t.Leaves[index].Hash())

	leafProof, err := t.inner.Prove(index)
	if err != nil {
		return nil, err
	}

	return &IndexedUpdateProof{
		LeafProof: IndexedMerkleProof{Siblings: leafProof},
		LeafIndex: index,
		PrevLeaf:  *prevLeaf,
	}, nil
}

func (t *IndexedMerkleTree) Insert(key *big.Int, value uint64) (*IndexedInsertionProof, error) {
	index := len(t.Leaves)
	lowIndex := t.GetLowIndex(key)

	prevLowLeaf := new(IndexedMerkleLeaf).Set(t.GetLeaf(lowIndex))
	if prevLowLeaf.Key.Cmp(key) == 0 {
		return nil, errors.New("key already exists")
	}

	newLowLeaf := new(IndexedMerkleLeaf).Set(&IndexedMerkleLeaf{
		Key:       prevLowLeaf.Key,
		Value:     prevLowLeaf.Value,
		NextIndex: index,
		NextKey:   new(big.Int).Set(key),
	})

	t.Leaves[lowIndex].Set(newLowLeaf)
	t.inner.updateLeaf(lowIndex, newLowLeaf.Hash())
	lowLeafProof, _, err := t.Prove(lowIndex)
	if err != nil {
		return nil, err
	}

	leaf := new(IndexedMerkleLeaf).Set(&IndexedMerkleLeaf{
		Key:       new(big.Int).Set(key),
		Value:     value,
		NextIndex: prevLowLeaf.NextIndex,
		NextKey:   prevLowLeaf.NextKey,
	})

	t.Leaves = append(t.Leaves, leaf)
	t.inner.updateLeaf(index, leaf.Hash())

	leafProof, _, err := t.Prove(index)
	if err != nil {
		return nil, err
	}

	return &IndexedInsertionProof{
		Index:        index,
		LowLeafProof: IndexedMerkleProof{Siblings: lowLeafProof},
		LeafProof:    IndexedMerkleProof{Siblings: leafProof},
		LowLeafIndex: lowIndex,
		PrevLowLeaf:  *prevLowLeaf,
	}, nil
}
