package models

import (
	"time"
)

type Deposit struct {
	ID                string
	DepositID         int64
	DepositHash       string
	RecipientSaltHash string
	TokenIndex        int64
	Amount            string // uint256
	CreatedAt         time.Time
}
