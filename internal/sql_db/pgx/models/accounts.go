package models

import (
	"time"

	"github.com/holiman/uint256"
)

type Account struct {
	ID        string
	AccountID uint256.Int
	SenderID  string
	CreatedAt time.Time
}
