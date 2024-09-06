package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

const base10 = 10

type LeafIndex = int

type IndexedMerkleLeaf struct {
	Key       *big.Int
	Value     uint64
	NextIndex LeafIndex
	NextKey   *big.Int
}

type SerializableIndexedMerkleLeaf struct {
	NextIndex LeafIndex `json:"nextIndex"`
	Key       string    `json:"key"`
	NextKey   string    `json:"nextKey"`
	Value     uint64    `json:"value"`
}

func (leaf *IndexedMerkleLeaf) MarshalJSON() ([]byte, error) {
	return json.Marshal(&SerializableIndexedMerkleLeaf{
		Key:       leaf.Key.String(),
		Value:     leaf.Value,
		NextIndex: leaf.NextIndex,
		NextKey:   leaf.NextKey.String(),
	})
}

func (leaf *IndexedMerkleLeaf) UnmarshalJSON(data []byte) error {
	aux := SerializableIndexedMerkleLeaf{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	leaf.Value = aux.Value
	leaf.NextIndex = aux.NextIndex

	key, ok := new(big.Int).SetString(aux.Key, base10)
	if !ok {
		return errors.New("invalid key")
	}
	leaf.Key = key

	nextKey, ok := new(big.Int).SetString(aux.NextKey, base10)
	if !ok {
		return errors.New("invalid next key")
	}
	leaf.NextKey = nextKey

	return nil
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
	Siblings []*goldenposeidon.PoseidonHashOut
}

func (proof *IndexedMerkleProof) MarshalJSON() ([]byte, error) {
	return json.Marshal(proof.Siblings)
}

func (proof *IndexedMerkleProof) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &proof.Siblings)
}

func NewDummyIndexedMerkleProof(height uint8) *IndexedMerkleProof {
	siblings := make([]*goldenposeidon.PoseidonHashOut, height)
	for i := range siblings {
		siblings[i] = goldenposeidon.NewPoseidonHashOut()
	}

	return &IndexedMerkleProof{Siblings: siblings}
}

func (proof *IndexedMerkleProof) IsDummy(height uint8) bool {
	dummyProof := NewDummyIndexedMerkleProof(height)
	for i := range proof.Siblings {
		if !proof.Siblings[i].Equal(dummyProof.Siblings[i]) {
			return false
		}
	}

	return true
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
		return fmt.Errorf("invalid root: %s != %s", computedRoot.String(), root.String())
	}

	return nil
}

type IndexedInsertionProof struct {
	Index        LeafIndex           `json:"index"`
	LowLeafProof *IndexedMerkleProof `json:"lowLeafProof"`
	LeafProof    *IndexedMerkleProof `json:"leafProof"`
	LowLeafIndex LeafIndex           `json:"lowLeafIndex"`
	PrevLowLeaf  *IndexedMerkleLeaf  `json:"prevLowLeaf"`
}

func (proof *IndexedInsertionProof) GetNewRoot(
	key *big.Int,
	value uint64,
	prevRoot *PoseidonHashOut,
) (*PoseidonHashOut, error) {
	// Ensure key > prevLowLeaf.Key
	if proof.PrevLowLeaf.Key.Cmp(key) != -1 {
		return nil, errors.New("key is not lower-bounded")
	}

	// Ensure prevLowLeaf.NextKey == 0 or key < prevLowLeaf.NextKey
	if proof.PrevLowLeaf.NextKey.Cmp(big.NewInt(0)) != 0 && proof.PrevLowLeaf.NextKey.Cmp(key) != 1 {
		return nil, errors.New("key is not upper-bounded")
	}

	err := proof.LowLeafProof.Verify(proof.PrevLowLeaf, proof.LowLeafIndex, prevRoot)
	if err != nil {
		return nil, errors.Join(ErrInvalidPrevRoot, err)
	}

	newLowLeaf := IndexedMerkleLeaf{
		NextIndex: proof.Index,
		NextKey:   key,
		Key:       proof.PrevLowLeaf.Key,
		Value:     proof.PrevLowLeaf.Value,
	}
	rootAfterUpdatedPrevLeaf := proof.LowLeafProof.GetRoot(&newLowLeaf, proof.LowLeafIndex)
	err = proof.LeafProof.Verify(
		new(IndexedMerkleLeaf).EmptyLeaf(),
		proof.Index,
		rootAfterUpdatedPrevLeaf,
	)
	if err != nil {
		return nil, errors.Join(ErrInvalidRootAfterUpdatedPrevLeaf, err)
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
	LeafProof IndexedMerkleProof `json:"leafProof"`
	LeafIndex LeafIndex          `json:"leafIndex"`
	PrevLeaf  IndexedMerkleLeaf  `json:"prevLeaf"`
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

func (t *IndexedMerkleTree) Set(other *IndexedMerkleTree) *IndexedMerkleTree {
	t.Leaves = make([]*IndexedMerkleLeaf, len(other.Leaves))
	for i, leaf := range other.Leaves {
		t.Leaves[i] = new(IndexedMerkleLeaf).Set(leaf)
	}

	t.inner = new(PoseidonMerkleTree).Set(other.inner)

	return t
}

type IndexedMembershipProof struct {
	IsIncluded bool               `json:"isIncluded"`
	LeafProof  IndexedMerkleProof `json:"leafProof"`
	LeafIndex  LeafIndex          `json:"leafIndex"`
	Leaf       IndexedMerkleLeaf  `json:"leaf"`
}

func (proof *IndexedMembershipProof) Verify(key *big.Int, root *PoseidonHashOut) error {
	err := proof.LeafProof.Verify(&proof.Leaf, proof.LeafIndex, root)
	if err != nil {
		return err
	}

	if proof.IsIncluded {
		if proof.Leaf.Key.Cmp(key) != 0 {
			return errors.New("key mismatch")
		}
	} else {
		if proof.Leaf.Key.Cmp(key) != -1 {
			return errors.New("key is not lower-bounded")
		}

		if proof.Leaf.NextKey.Cmp(big.NewInt(0)) != 0 && proof.Leaf.NextKey.Cmp(key) != 1 {
			return errors.New("key is not upper-bounded")
		}
	}

	return nil
}

func NewIndexedMerkleTree(height uint8, zeroHash *goldenposeidon.PoseidonHashOut) (*IndexedMerkleTree, error) {
	tree, err := NewPoseidonMerkleTree(height, zeroHash)
	if err != nil {
		return nil, err
	}

	defaultLeaf := new(IndexedMerkleLeaf).EmptyLeaf()
	defaultLeafHash := defaultLeaf.Hash()
	tree.updateLeaf(0, defaultLeafHash)

	return &IndexedMerkleTree{
		Leaves: []*IndexedMerkleLeaf{defaultLeaf},
		inner:  tree,
	}, nil
}

func (t *IndexedMerkleTree) GetRoot() *PoseidonHashOut {
	root := t.inner.GetRoot()

	return root
}

func (t *IndexedMerkleTree) GetLeaf(index LeafIndex) *IndexedMerkleLeaf {
	return t.Leaves[index]
}

func (t *IndexedMerkleTree) Prove(index LeafIndex) (proof *IndexedMerkleProof, root *PoseidonHashOut, err error) {
	innerProof, err := t.inner.Prove(index)
	if err != nil {
		return nil, nil, err
	}

	root = t.GetRoot()

	proof = &IndexedMerkleProof{Siblings: innerProof.Siblings}
	return proof, root, err
}

func (t *IndexedMerkleTree) ProveMembership(key *big.Int) (membership_proof *IndexedMembershipProof, root *PoseidonHashOut, err error) {
	lowIndex := t.GetLowIndex(key)
	lowLeaf := t.GetLeaf(lowIndex)
	leafProof, root, err := t.Prove(lowIndex)
	// fmt.Printf("lowIndex: %d\n", lowIndex)
	// fmt.Printf("lowLeaf.Key: %v, key: %s\n", lowLeaf.Key, key)
	// fmt.Printf("leafProof: %v\n", leafProof)
	if err != nil {
		return nil, nil, err
	}

	membership_proof = &IndexedMembershipProof{
		IsIncluded: lowLeaf.Key.Cmp(key) == 0,
		LeafProof:  *leafProof,
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
		LeafProof: IndexedMerkleProof{Siblings: leafProof.Siblings},
		LeafIndex: index,
		PrevLeaf:  *prevLeaf,
	}, nil
}

func (t *IndexedMerkleTree) Insert(key *big.Int, value uint64) (*IndexedInsertionProof, error) {
	index := len(t.Leaves)
	lowIndex := t.GetLowIndex(key)

	prevLowLeaf := new(IndexedMerkleLeaf).Set(t.GetLeaf(lowIndex))
	// fmt.Printf("lowIndex: %d\n", lowIndex)
	// fmt.Printf("prevLowLeaf.Key: %v, key: %s\n", prevLowLeaf, key)
	if prevLowLeaf.Key.Cmp(key) == 0 {
		return nil, errors.New("key already exists")
	}

	prevRoot := t.GetRoot()

	newLowLeaf := new(IndexedMerkleLeaf).Set(&IndexedMerkleLeaf{
		Key:       prevLowLeaf.Key,
		Value:     prevLowLeaf.Value,
		NextIndex: index,
		NextKey:   new(big.Int).Set(key),
	})

	t.Leaves[lowIndex].Set(newLowLeaf)
	t.inner.updateLeaf(lowIndex, newLowLeaf.Hash())
	lowLeafProof, rootAfterUpdatedPrevLeaf, err := t.Prove(lowIndex)
	if err != nil {
		return nil, err
	}

	// debug
	err = lowLeafProof.Verify(
		newLowLeaf,
		lowIndex,
		rootAfterUpdatedPrevLeaf,
	)
	if err != nil {
		fmt.Println("Fail to verify new low leaf")
		panic(err)
	}
	err = lowLeafProof.Verify(
		prevLowLeaf,
		lowIndex,
		prevRoot,
	)
	if err != nil {
		fmt.Println("Fail to verify old low leaf")
		panic(err)
	}

	leaf := new(IndexedMerkleLeaf).Set(&IndexedMerkleLeaf{
		Key:       new(big.Int).Set(key),
		Value:     value,
		NextIndex: prevLowLeaf.NextIndex,
		NextKey:   prevLowLeaf.NextKey,
	})

	t.Leaves = append(t.Leaves, leaf)
	t.inner.updateLeaf(index, leaf.Hash())

	leafProof, newRoot, err := t.Prove(index)
	if err != nil {
		return nil, err
	}

	// debug
	err = leafProof.Verify(
		leaf,
		index,
		newRoot,
	)
	if err != nil {
		fmt.Println("Fail to verify new")
		panic(err)
	}
	err = leafProof.Verify(
		new(IndexedMerkleLeaf).EmptyLeaf(),
		index,
		rootAfterUpdatedPrevLeaf,
	)
	if err != nil {
		fmt.Println("Fail to verify old")
		panic(err)
	}

	return &IndexedInsertionProof{
		Index:        index,
		LowLeafProof: lowLeafProof,
		LeafProof:    leafProof,
		LowLeafIndex: lowIndex,
		PrevLowLeaf:  prevLowLeaf,
	}, nil
}
