package tree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"intmax2-node/internal/finite_field"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const DEPOSIT_TREE_HEIGHT uint8 = 32

type DepositLeaf struct {
	RecipientSaltHash [numHashBytes]byte `json:"recipientSaltHash"`
	TokenIndex        uint32             `json:"tokenIndex"`
	Amount            *big.Int           `json:"amount"`
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
	return crypto.Keccak256Hash(dd.Marshal())
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

// pub fn to_u32_vec(&self) -> Vec<u32> {
// 	let vec = vec![
// 		self.pubkey_salt_hash.to_u32_vec(),
// 		vec![self.token_index],
// 		self.amount.to_u32_vec(),
// 	]
// 	.concat();
// 	vec
// }

// pub fn poseidon_hash(&self) -> PoseidonHashOut {
// 	PoseidonHashOut::hash_inputs_u32(&self.to_u32_vec())
// }

func (dd *DepositLeaf) ToFieldElementSlice() []ffg.Element {
	buf := finite_field.NewBuffer(make([]ffg.Element, 0))

	const int32Key = 32
	finite_field.WriteFixedSizeBytes(buf, dd.RecipientSaltHash[:], int32Key)
	finite_field.WriteUint32(buf, dd.TokenIndex)
	amountUint256 := new(intMaxTypes.Uint256).FromBigInt(dd.Amount)
	finite_field.WriteFixedSizeBytes(buf, amountUint256.Bytes(), int32Key)

	return buf.Inner()
}

func (dd *DepositLeaf) Nullifier() *PoseidonHashOut {
	return intMaxGP.HashNoPad(dd.ToFieldElementSlice())
}

type DepositTree struct {
	Leaves []*DepositLeaf
	inner  *KeccakMerkleTree
}

func (t *DepositTree) Set(other *DepositTree) *DepositTree {
	t.Leaves = make([]*DepositLeaf, len(other.Leaves))
	copy(t.Leaves, other.Leaves)
	t.inner = new(KeccakMerkleTree).Set(other.inner)

	return t
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

func (t *DepositTree) GetCurrentRootCountAndSiblings() (root common.Hash, nextIndex uint32, siblings *KeccakMerkleProof) {
	return t.inner.GetCurrentRootCountAndSiblings()
}

func (t *DepositTree) AddLeaf(index uint32, leaf DepositLeaf) (root [numHashBytes]byte, err error) {
	leafHash := leaf.Hash()
	root, err = t.inner.AddLeaf(index, leafHash)
	if err != nil {
		return [numHashBytes]byte{}, err
	}

	if int(index) != len(t.Leaves) {
		return [numHashBytes]byte{}, errors.New("index is not equal to the length of deposit leaves")
	}

	t.Leaves = append(t.Leaves, new(DepositLeaf).Set(&leaf))

	return root, nil
}

func (t *DepositTree) ComputeMerkleProof(index uint32, leaves [][numHashBytes]byte) (proof *KeccakMerkleProof, root common.Hash, err error) {
	return t.inner.ComputeMerkleProof(index, leaves)
}

func (t *DepositTree) Prove(index uint32) (proof *KeccakMerkleProof, root common.Hash, err error) {
	leaves := make([][32]byte, 0)
	for _, depositLeaf := range t.Leaves {
		leaves = append(leaves, [32]byte(depositLeaf.Hash()))
	}

	return t.inner.ComputeMerkleProof(index, leaves)
}
