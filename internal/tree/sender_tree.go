package tree

import (
	"errors"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/iden3/go-iden3-crypto/ffg"
)

type SenderLeaf struct {
	Sender  *intMaxTypes.Uint256
	IsValid bool
}

func (l *SenderLeaf) Set(leaf *SenderLeaf) *SenderLeaf {
	return &SenderLeaf{
		Sender:  leaf.Sender,
		IsValid: leaf.IsValid,
	}
}

func (l *SenderLeaf) ToFieldElementSlice() []ffg.Element {
	isValid := new(ffg.Element).SetUint64(0)
	if l.IsValid {
		isValid = new(ffg.Element).SetUint64(1)
	}
	return append(l.Sender.ToFieldElementSlice(), *isValid)
}

func (l *SenderLeaf) Hash() *PoseidonHashOut {
	return intMaxGP.HashNoPad(l.ToFieldElementSlice())
}

type SenderTree struct {
	Leaves []*SenderLeaf
	inner  *PoseidonIncrementalMerkleTree
}

func NewSenderTree(
	height uint8,
	initialLeaves []*SenderLeaf,
) (*SenderTree, error) {
	if len(initialLeaves) == 1<<height {
		return nil, errors.New("initialLeaves length is equal to 2^height")
	}

	initialLeafHashes := make([]*PoseidonHashOut, len(initialLeaves))
	for key := range initialLeaves {
		initialLeafHashes[key] = initialLeaves[key].Hash()
	}

	zeroHash := intMaxGP.HashNoPad([]ffg.Element{}) // unused
	t, err := NewPoseidonIncrementalMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, errors.Join(ErrNewPoseidonMerkleTreeFail, err)
	}

	leaves := make([]*SenderLeaf, len(initialLeaves))
	for key := range initialLeaves {
		leaves[key] = new(SenderLeaf).Set(initialLeaves[key])
	}

	return &SenderTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}
