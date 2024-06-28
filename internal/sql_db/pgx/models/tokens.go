package models

import (
	"database/sql"
	"time"

	"github.com/holiman/uint256"
)

type Token struct {
	ID           string
	TokenIndex   string
	TokenAddress sql.NullString
	TokenID      uint256.Int
	CreatedAt    time.Time
}
