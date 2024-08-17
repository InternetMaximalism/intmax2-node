package tree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const DEPOSIT_TREE_HEIGHT uint8 = 32

type DepositLeaf struct {
	RecipientSaltHash [numHashBytes]byte
	TokenIndex        uint32
	Amount            *big.Int
}

func (dd *DepositLeaf) Set(deposit *DepositLeaf) *DepositLeaf {
	dd.RecipientSaltHash = deposit.RecipientSaltHash
	dd.TokenIndex = deposit.TokenIndex
	dd.Amount = deposit.Amount
	return dd
}

func (dd *DepositLeaf) SetZero() *DepositLeaf {
	dd.RecipientSaltHash = [numHashBytes]byte{}
	dd.TokenIndex = 0
	dd.Amount = big.NewInt(0)
	return dd
}

func (dd *DepositLeaf) Marshal() []byte {
	const (
		int4Key  = 4
		int31Key = 31
		int32Key = 32
	)

	tokenIndexBytes := make([]byte, int4Key)
	binary.BigEndian.PutUint32(tokenIndexBytes, dd.TokenIndex)
	amountBytes := intMaxTypes.BigIntToBytes32BeArray(dd.Amount)

	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(dd.RecipientSaltHash[:])
	err := binary.Write(buf, binary.BigEndian, dd.TokenIndex)
	if err != nil {
		panic(err)
	}
	buf.Write(amountBytes[:])

	return buf.Bytes()
}

func (dd *DepositLeaf) Hash() common.Hash {
	packed := dd.Marshal()
	fmt.Printf("packed: %x\n", packed)

	res := crypto.Keccak256Hash(dd.Marshal())
	fmt.Printf("hash: %x\n", res)
	return res
}

func (dd *DepositLeaf) Equal(other *DepositLeaf) bool {
	switch {
	case dd.RecipientSaltHash != other.RecipientSaltHash,
		dd.TokenIndex != other.TokenIndex,
		dd.Amount.Cmp(other.Amount) != 0:
		return false
	default:
		return true
	}
}

type DepositTree struct {
	Leaves []*DepositLeaf
	inner  *KeccakMerkleTree
}

func NewDepositTree(height uint8) (*DepositTree, error) {
	return NewDepositTreeWithInitialLeaves(height, nil)
}

func NewDepositTreeWithInitialLeaves(height uint8, initialLeaves []*DepositLeaf) (*DepositTree, error) {
	zeroHash := new(DepositLeaf).SetZero().Hash()

	initialLeafHashes := make([][32]byte, len(initialLeaves))
	for i, leaf := range initialLeaves {
		initialLeafHashes[i] = leaf.Hash()
	}

	t, err := NewKeccakMerkleTree(height, initialLeafHashes, zeroHash)
	if err != nil {
		return nil, err
	}

	leaves := make([]*DepositLeaf, len(initialLeaves))
	copy(leaves, initialLeaves)

	return &DepositTree{
		Leaves: leaves,
		inner:  t,
	}, nil
}

func (t *DepositTree) BuildMerkleRoot(leaves [][numHashBytes]byte) (common.Hash, error) {
	return t.inner.BuildMerkleRoot(leaves)
}

func (t *DepositTree) GetCurrentRootCountAndSiblings() (root common.Hash, nextIndex uint32, siblings [][numHashBytes]byte) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *DepositTree) AddLeaf(index uint32, leaf DepositLeaf) (root [numHashBytes]byte, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return [numHashBytes]byte{}, err
	}

	if int(index) != len(t.Leaves) {
		return [numHashBytes]byte{}, errors.New("index is not equal to the length of leaves")
	}

	t.Leaves = append(t.Leaves, new(DepositLeaf).Set(&leaf))

	return root, nil
}

func (t *DepositTree) ComputeMerkleProof(index uint32, leaves [][numHashBytes]byte) (siblings [][numHashBytes]byte, root common.Hash, err error) {
	return t.inner.ComputeMerkleProof(index, leaves)
}
