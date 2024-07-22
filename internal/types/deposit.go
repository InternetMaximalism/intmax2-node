package types

import (
	"encoding/binary"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const numHashBytes = 32

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
	amountBytes := make([]byte, int32Key)
	for i, v := range dd.Amount.Bytes() {
		amountBytes[int31Key-i] = v
	}

	return append(
		append(dd.RecipientSaltHash[:], tokenIndexBytes...),
		amountBytes...,
	)
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
