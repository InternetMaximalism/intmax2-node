package models

import (
	"time"

	"github.com/holiman/uint256"
)

type Token struct {
	ID           string
	TokenIndex   string
	TokenAddress string
	TokenID      uint256.Int
	CreatedAt    time.Time
}
