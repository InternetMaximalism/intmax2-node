package tree

import (
	"errors"
	"fmt"
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

	return append([]ffg.Element{*isInsufficient}, l.Amount.ToFieldElementSlice()[:]...)
}

func (l *AssetLeaf) Hash() *PoseidonHashOut {
	return intMaxGP.HashNoPad(l.ToFieldElementSlice())
}

// /// Substracts the given amount from the balance. If the balance is insufficient, the balance
// /// becomes zero and the `is_insufficient` flag is set to true.
// pub fn sub(&self, amount: U256) -> Self {
//     let is_insufficient = (self.amount < amount) || self.is_insufficient;
//     // If the balance is insufficient, substract the entire balance to prevent
//     // underflow.
//     let substract_amount = if is_insufficient { self.amount } else { amount };
//     let amount = self.amount - substract_amount;
//     Self {
//         is_insufficient,
//         amount,
//     }
// }

// /// Adds the given amount to the balance. `is_insufficient` flag is not changed.
// pub fn add(&self, amount: U256) -> Self {
//     Self {
//         is_insufficient: self.is_insufficient,
//         amount: self.amount + amount,
//     }
// }

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

type AssetMerkleProof struct {
	Siblings []*PoseidonHashOut `json:"siblings"`
}

func (proof *AssetMerkleProof) GetRoot(
	leaf *AssetLeaf,
	index uint32,
) *PoseidonHashOut {
	merkleProof := MerkleProof{
		Siblings: proof.Siblings,
	}
	root := merkleProof.GetRoot(
		leaf.Hash(),
		int(index),
	)

	return root
}

func (proof *AssetMerkleProof) Verify(
	leaf *AssetLeaf,
	index uint32,
	root *PoseidonHashOut,
) error {
	merkleProof := MerkleProof{
		Siblings: proof.Siblings,
	}
	return merkleProof.Verify(
		leaf.Hash(),
		int(index),
		root,
	)
}

type AssetTree struct {
	Leaves []*AssetLeaf
	inner  *PoseidonIncrementalMerkleTree
}

func NewAssetTree(
	height uint8,
	initialLeaves []*AssetLeaf,
	zeroHash *PoseidonHashOut,
) (*AssetTree, error) {
	initialLeafHashes := make([]*PoseidonHashOut, len(initialLeaves))
	for key := range initialLeaves {
		initialLeafHashes[key] = initialLeaves[key].Hash()
	}

	t, err := NewPoseidonIncrementalMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, errors.Join(ErrNewPoseidonMerkleTreeFail, err)
	}

	leaves := make([]*AssetLeaf, len(initialLeaves))
	for key := range initialLeaves {
		leaves[key] = new(AssetLeaf).Set(initialLeaves[key])
	}

	return &AssetTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *AssetTree) BuildMerkleRoot(leaves []*AssetLeaf) (root *PoseidonHashOut, err error) {
	leafHashes := make([]*PoseidonHashOut, len(leaves))
	for key := range leaves {
		leafHashes[key] = leaves[key].Hash()
	}

	return t.inner.BuildMerkleRoot(leafHashes)
}

// GetCurrentRootCountAndSiblings returns the latest root, count and siblings
func (t *AssetTree) GetCurrentRootCountAndSiblings() (root PoseidonHashOut, count uint64, siblings []*PoseidonHashOut) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *AssetTree) AddLeaf(index uint32, leaf *AssetLeaf) (root *PoseidonHashOut, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(uint64(index), leafHash)
	if err != nil {
		return nil, errors.Join(ErrAddLeafFail, err)
	}

	if int(index) != len(t.Leaves) {
		return nil, errors.Join(ErrLeafInputIndexInvalid, errors.New("asset tree AddLeaf"))
	}
	t.Leaves = append(t.Leaves, new(AssetLeaf).Set(leaf))

	return root, nil
}

func (t *AssetTree) ComputeMerkleProof(
	index uint32,
) (siblings []*PoseidonHashOut, root PoseidonHashOut, err error) {
	leaves := make([]*PoseidonHashOut, 1<<t.inner.height)
	for i, leaf := range t.Leaves {
		leaves[i] = leaf.Hash()
	}
	for i := len(t.Leaves); i < len(leaves); i++ {
		leaves[i] = t.inner.zeroHashes[0]
	}

	return t.inner.ComputeMerkleProof(uint64(index), leaves)
}

func (t *AssetTree) GetLeaf(index uint32) *AssetLeaf {
	if index >= uint32(len(t.Leaves)) {
		return new(AssetLeaf).SetDefault()
	}

	return t.Leaves[index]
}

func (t *AssetTree) GetRoot() *PoseidonHashOut {
	root, _, _ := t.inner.GetCurrentRootCountAndSiblings()
	return &root
}

func (t *AssetTree) UpdateLeaf(index uint32, leaf *AssetLeaf) (root *PoseidonHashOut, err error) {
	if index >= uint32(len(t.Leaves)) {
		fmt.Printf("index: %d, len(t.Leaves): %d\n", index, len(t.Leaves))
		return nil, errors.Join(ErrLeafInputIndexInvalid, errors.New("asset tree UpdateLeaf"))
	}

	t.Leaves[index] = leaf
	return t.inner.UpdateLeaf(uint64(index), leaf.Hash())
}

func (t *AssetTree) Prove(index uint32) (proof *AssetMerkleProof, root PoseidonHashOut, err error) {
	proof = new(AssetMerkleProof)
	proof.Siblings, root, err = t.ComputeMerkleProof(index)
	if err != nil {
		var ErrComputeMerkleProofFail = errors.New("compute merkle proof fail")
		return nil, PoseidonHashOut{}, errors.Join(ErrComputeMerkleProofFail, err)
	}

	return proof, root, nil
}
