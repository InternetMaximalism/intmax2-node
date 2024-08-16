package tree

import (
	"errors"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type IndexedMerkleLeaf struct {
	NextIndex uint64   `json:"nextIndex"`
	Key       *big.Int `json:"key"`
	NextKey   *big.Int `json:"nextKey"`
	Value     uint64   `json:"value"`
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
	err := finite_field.WriteUint64(res, leaf.NextIndex)
	if err != nil {
		panic(err)
	}
	key := new(intMaxTypes.Uint256).FromBigInt(leaf.Key).ToFieldElementSlice()
	for _, limb := range key {
		finite_field.WriteGoldilocksField(res, &limb)
	}
	nextKey := new(intMaxTypes.Uint256).FromBigInt(leaf.NextKey).ToFieldElementSlice()
	for _, limb := range nextKey {
		finite_field.WriteGoldilocksField(res, &limb)
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
func (proof *IndexedMerkleProof) GetRoot(leaf *IndexedMerkleLeaf, index int) *goldenposeidon.PoseidonHashOut {
	height := len(proof.Siblings)
	if index >= 1<<int(height) {
		panic("index out of bounds")
	}
	nodeIndex := 1<<int(height) + index
	h := new(PoseidonHashOut).Set(leaf.Hash())

	for i := 0; i < height; i++ {
		sibling := proof.Siblings[i]
		if nodeIndex&1 == 1 {
			h = goldenposeidon.Compress(sibling, h)
		} else {
			h = goldenposeidon.Compress(h, sibling)
		}
		nodeIndex = nodeIndex >> 1
	}
	if nodeIndex != 1 {
		panic("invalid nodeIndex")
	}

	return h
}

// TODO: leaf is *BlockHashLeaf?
func (proof *IndexedMerkleProof) Verify(leaf *IndexedMerkleLeaf, index int, root *goldenposeidon.PoseidonHashOut) error {
	computedRoot := proof.GetRoot(leaf, index)
	if !computedRoot.Equal(root) {
		return errors.New("invalid root")
	}

	return nil
}

type IndexedInsertionProof struct {
	Index        uint64
	LowLeafProof IndexedMerkleProof
	LeafProof    IndexedMerkleProof
	LowLeafIndex uint64
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

	err := proof.LowLeafProof.Verify(&proof.PrevLowLeaf, int(proof.LowLeafIndex), prevRoot)
	if err != nil {
		return nil, err
	}

	newLowLeaf := IndexedMerkleLeaf{
		NextIndex: proof.Index,
		NextKey:   key,
		Key:       proof.PrevLowLeaf.Key,
		Value:     proof.PrevLowLeaf.Value,
	}
	tempRoot := proof.LowLeafProof.GetRoot(&newLowLeaf, int(proof.LowLeafIndex))
	err = proof.LeafProof.Verify(
		new(IndexedMerkleLeaf).EmptyLeaf(),
		int(proof.Index),
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

	return proof.LeafProof.GetRoot(&leaf, int(proof.Index)), nil
}

func (proof *IndexedInsertionProof) ConditionalGetNewRoot(condition bool, key *big.Int, value uint64, prevRoot *PoseidonHashOut) (*PoseidonHashOut, error) {
	if condition {
		return proof.GetNewRoot(key, value, prevRoot)
	}

	return prevRoot, nil
}

type IndexedUpdateProof struct {
	LeafProof IndexedMerkleProof
	LeafIndex uint64
	PrevLeaf  IndexedMerkleLeaf
}

func (proof *IndexedUpdateProof) GetNewRoot(key *big.Int, prevValue uint64, newValue uint64, prevRoot *PoseidonHashOut) (*PoseidonHashOut, error) {
	if proof.PrevLeaf.Value != prevValue {
		return nil, errors.New("value mismatch")
	}

	if proof.PrevLeaf.Key.Cmp(key) != 0 {
		return nil, errors.New("key mismatch")
	}

	err := proof.LeafProof.Verify(&proof.PrevLeaf, int(proof.LeafIndex), prevRoot)
	if err != nil {
		return nil, err
	}

	newLeaf := IndexedMerkleLeaf{
		Value:     newValue,
		NextIndex: proof.PrevLeaf.NextIndex,
		Key:       proof.PrevLeaf.Key,
		NextKey:   proof.PrevLeaf.NextKey,
	}
	return proof.LeafProof.GetRoot(&newLeaf, int(proof.LeafIndex)), nil
}

func (proof *IndexedUpdateProof) Verify(key *big.Int, prevValue uint64, newValue uint64, prevRoot *PoseidonHashOut, newRoot *PoseidonHashOut) error {
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

func (t *IndexedMerkleTree) GetLeaf(index uint64) *IndexedMerkleLeaf {
	return t.Leaves[index]
}

func (t *IndexedMerkleTree) Prove(index uint64) ([]*PoseidonHashOut, PoseidonHashOut, error) {
	siblings, err := t.inner.Prove(int(index))
	if err != nil {
		return nil, goldenposeidon.PoseidonHashOut{}, err
	}

	root := t.GetRoot()

	return siblings, root, err
}

func (t *IndexedMerkleTree) GetLowIndex(key *big.Int) (int, error) {
	validLeafCandidates := make([]int, 0)
	for i, leaf := range t.Leaves {
		// key > leaf.Key
		isValidLowerLimit := key.Cmp(leaf.Key) == 1
		// key < leaf.NextKey || leaf.NextKey == 0
		isValidUpperLimit := key.Cmp(leaf.NextKey) == -1 || leaf.NextKey.Cmp(big.NewInt(0)) == 0
		if isValidLowerLimit && isValidUpperLimit {
			validLeafCandidates = append(validLeafCandidates, i)
		}
	}

	if len(validLeafCandidates) == 0 {
		var ErrKeyAlreadyExists = errors.New("key already exists")
		return -1, ErrKeyAlreadyExists
	}
	if len(validLeafCandidates) > 1 {
		var ErrTooManyCandidates = errors.New("too many candidates")
		panic(ErrTooManyCandidates)
	}

	return validLeafCandidates[0], nil
}

func (t *IndexedMerkleTree) index(key *big.Int) int {
	validLeafCandidates := make([]int, 0)
	for i, leaf := range t.Leaves {
		if key.Cmp(leaf.Key) == 0 {
			validLeafCandidates = append(validLeafCandidates, i)
		}
	}

	if len(validLeafCandidates) == 0 {
		return -1
	}
	if len(validLeafCandidates) > 1 {
		var ErrTooManyCandidates = errors.New("too many candidates")
		panic(ErrTooManyCandidates)
	}

	return validLeafCandidates[0]
}

func (t *IndexedMerkleTree) Update(key *big.Int, value uint64) (int, error) {
	if key.Cmp(big.NewInt(0)) == -1 {
		return -1, errors.New("key is negative")
	}
	const uint256Key uint = 256
	if key.Cmp(new(big.Int).Lsh(big.NewInt(1), uint256Key)) == 1 {
		return -1, errors.New("key is too large")
	}

	index := t.index(key)
	if index == -1 {
		return -1, errors.New("key doesn't exist")
	}

	leaf := t.Leaves[index]
	leaf.Value = value
	t.Leaves[index] = leaf
	t.inner.updateLeaf(index, leaf.Hash())

	return index, nil
}
