package models

import (
	"time"
)

type Deposit struct {
	ID                string
	DepositID         uint32
	DepositHash       string
	RecipientSaltHash string
	TokenIndex        uint32
	Amount            string // uint256
	CreatedAt         time.Time
}
