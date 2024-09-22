package tree

import (
	"encoding/binary"
	"errors"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

const ASSET_TREE_HEIGHT = 32

type AssetLeaf struct {
	IsInsufficient bool
	Amount         *intMaxTypes.Uint256
}

func (l *AssetLeaf) Marshal() []byte {
	amount := l.Amount.Bytes()
	isInsufficient := byte(0)
	if l.IsInsufficient {
		isInsufficient = byte(1)
	}

	return append([]byte{isInsufficient}, amount...)
}

func (l *AssetLeaf) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return errors.New("invalid data length")
	}

	l.IsInsufficient = data[0] == 1
	l.Amount = new(intMaxTypes.Uint256).FromBytes(data[1:])

	return nil
}

func (l *AssetLeaf) Set(leaf *AssetLeaf) *AssetLeaf {
	return &AssetLeaf{
		IsInsufficient: leaf.IsInsufficient,
		Amount:         leaf.Amount,
	}
}

func (l *AssetLeaf) SetDefault() *AssetLeaf {
	return &AssetLeaf{
		IsInsufficient: false,
		Amount:         new(intMaxTypes.Uint256).FromBigInt(big.NewInt(0)),
	}
}

func (l *AssetLeaf) ToFieldElementSlice() []ffg.Element {
	isInsufficient := new(ffg.Element).SetUint64(0)
	if l.IsInsufficient {
		isInsufficient = new(ffg.Element).SetUint64(1)
	}

	return append([]ffg.Element{*isInsufficient}, l.Amount.ToFieldElementSlice()...)
}

func (l *AssetLeaf) Hash() *PoseidonHashOut {
	return intMaxGP.HashNoPad(l.ToFieldElementSlice())
}

func (l *AssetLeaf) Add(amount *big.Int) *AssetLeaf {
	return &AssetLeaf{
		IsInsufficient: l.IsInsufficient,
		Amount:         l.Amount.Add(new(intMaxTypes.Uint256).FromBigInt(amount)),
	}
}

func (l *AssetLeaf) Sub(amount *big.Int) *AssetLeaf {
	isInsufficient := l.IsInsufficient || l.Amount.BigInt().Cmp(amount) < 0
	subAmount := l.Amount
	if !isInsufficient {
		subAmount = new(intMaxTypes.Uint256).FromBigInt(amount)
	}

	return &AssetLeaf{
		IsInsufficient: isInsufficient,
		Amount:         l.Amount.Sub(subAmount),
	}
}

// type AssetMerkleProof struct {
// 	Siblings []*PoseidonHashOut `json:"siblings"`
// }

type AssetMerkleProof = PoseidonMerkleProof

// func (proof *AssetMerkleProof) GetRoot(
// 	leaf *AssetLeaf,
// 	index uint32,
// ) *PoseidonHashOut {
// 	merkleProof := PoseidonMerkleProof{
// 		Siblings: proof.Siblings,
// 	}
// 	root := merkleProof.GetRoot(
// 		leaf.Hash(),
// 		int(index),
// 	)

// 	return root
// }

// func (proof *AssetMerkleProof) Verify(
// 	leaf *AssetLeaf,
// 	index uint32,
// 	root *PoseidonHashOut,
// ) error {
// 	merkleProof := PoseidonMerkleProof{
// 		Siblings: proof.Siblings,
// 	}
// 	return merkleProof.Verify(
// 		leaf.Hash(),
// 		int(index),
// 		root,
// 	)
// }

type AssetLeafEntry struct {
	TokenIndex uint32
	Leaf       *AssetLeaf
}

func (entry *AssetLeafEntry) Set(other *AssetLeafEntry) *AssetLeafEntry {
	entry.TokenIndex = other.TokenIndex
	entry.Leaf = new(AssetLeaf).Set(other.Leaf)

	return entry
}

func (entry *AssetLeafEntry) Marshal() []byte {
	tokenIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(tokenIndexBytes, entry.TokenIndex)
	leaf := entry.Leaf.Marshal()

	return append(tokenIndexBytes, leaf...)
}

func (entry *AssetLeafEntry) Unmarshal(data []byte) (*AssetLeafEntry, error) {
	if len(data) < 4 {
		return nil, errors.New("invalid data length")
	}

	entry.TokenIndex = binary.BigEndian.Uint32(data[:4])
	entry.Leaf = new(AssetLeaf)
	if err := entry.Leaf.Unmarshal(data[4:]); err != nil {
		return nil, err
	}

	return entry, nil
}

type AssetTree struct {
	Leaves map[uint32]*AssetLeaf
	inner  *PoseidonMerkleTree
}

func NewAssetTree(
	height uint8,
	initialLeaves []*AssetLeafEntry,
	zeroHash *PoseidonHashOut,
) (*AssetTree, error) {
	t, err := NewPoseidonMerkleTree(height, zeroHash)
	if err != nil {
		return nil, errors.Join(ErrNewPoseidonMerkleTreeFail, err)
	}

	for _, value := range initialLeaves {
		t.UpdateLeaf(int(value.TokenIndex), value.Leaf.Hash())
	}

	leaves := make(map[uint32]*AssetLeaf, len(initialLeaves))
	for _, value := range initialLeaves {
		leaves[value.TokenIndex] = new(AssetLeaf).Set(value.Leaf)
	}

	return &AssetTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *AssetTree) Set(other *AssetTree) *AssetTree {
	t.Leaves = make(map[uint32]*AssetLeaf, len(other.Leaves))
	for key := range other.Leaves {
		t.Leaves[key] = new(AssetLeaf).Set(other.Leaves[key])
	}

	t.inner = new(PoseidonMerkleTree).Set(other.inner)

	return t
}

// func (t *AssetTree) BuildMerkleRoot(leaves []*AssetLeaf) (root *PoseidonHashOut, err error) {
// 	leafHashes := make([]*PoseidonHashOut, len(leaves))
// 	for key := range leaves {
// 		leafHashes[key] = leaves[key].Hash()
// 	}

// 	return t.inner.BuildMerkleRoot(leafHashes)
// }

// // GetCurrentRootCountAndSiblings returns the latest root, count and siblings
// func (t *AssetTree) GetCurrentRootCountAndSiblings() (root PoseidonHashOut, count uint64, siblings []*PoseidonHashOut) {
// 	return t.inner.GetCurrentRootCountAndSiblings()
// }

// func (t *AssetTree) AddLeaf(index uint32, leaf *AssetLeaf) (root *PoseidonHashOut, err error) {
// 	leafHash := leaf.Hash()
// 	root, err = t.inner.AddLeaf(uint64(index), leafHash)
// 	if err != nil {
// 		return nil, errors.Join(ErrAddLeafFail, err)
// 	}

// 	if int(index) != len(t.Leaves) {
// 		return nil, errors.Join(ErrAssetLeafInputIndexInvalid, errors.New("asset tree AddLeaf"))
// 	}
// 	t.Leaves = append(t.Leaves, new(AssetLeaf).Set(leaf))

// 	return root, nil
// }

// func (t *AssetTree) ComputeMerkleProof(
// 	index uint32,
// ) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
// 	leaves := make([]*PoseidonHashOut, len(t.Leaves))
// 	for i, leaf := range t.Leaves {
// 		leaves[i] = leaf.Hash()
// 	}
// 	// for i := len(t.Leaves); i < len(leaves); i++ {
// 	// 	leaves[i] = t.inner.zeroHashes[0]
// 	// }

// 	return t.inner.ComputeMerkleProof(uint64(index), leaves)
// }

func (t *AssetTree) GetLeaf(index uint32) *AssetLeaf {
	if leaf, ok := t.Leaves[index]; ok {
		return leaf
	}

	return new(AssetLeaf).SetDefault()
}

func (t *AssetTree) GetRoot() *PoseidonHashOut {
	return t.inner.GetRoot()
}

func (t *AssetTree) UpdateLeaf(tokenIndex uint32, leaf *AssetLeaf) (root *PoseidonHashOut, err error) {
	t.Leaves[tokenIndex] = leaf
	t.inner.UpdateLeaf(int(tokenIndex), leaf.Hash())

	return t.inner.GetRoot(), nil
}

func (t *AssetTree) Prove(tokenIndex uint32) (proof AssetMerkleProof, root PoseidonHashOut, err error) {
	proof, err = t.inner.Prove(int(tokenIndex))
	if err != nil {
		var ErrComputeMerkleProofFail = errors.New("compute merkle proof fail")
		return AssetMerkleProof{}, PoseidonHashOut{}, errors.Join(ErrComputeMerkleProofFail, err)
	}

	return proof, root, nil
}
