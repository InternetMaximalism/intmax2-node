package models

import (
	"database/sql"
	"time"

	"github.com/holiman/uint256"
)

type Deposit struct {
	ID                            string
	DepositID                     int64
	DepositIndex                  *int64
	DepositHash                   string
	RecipientSaltHash             string
	TokenIndex                    int64
	Amount                        uint256.Int
	Sender                        sql.NullString
	BlockNumberAfterDepositIndex  uint32
	BlockNumberBeforeDepositIndex uint32
	IsSync                        uint32
	CreatedAt                     time.Time
}
