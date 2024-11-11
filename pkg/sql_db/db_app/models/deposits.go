package models

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const int32Key = 32

type Deposit struct {
	ID                            string
	DepositID                     uint32
	DepositIndex                  *uint32
	DepositHash                   common.Hash
	RecipientSaltHash             [int32Key]byte
	TokenIndex                    uint32
	Amount                        *big.Int
	Sender                        string
	BlockNumberAfterDepositIndex  uint32
	BlockNumberBeforeDepositIndex uint32
	IsSync                        bool
	CreatedAt                     time.Time
}
