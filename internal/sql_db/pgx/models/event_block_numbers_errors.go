package models

import (
	"encoding/json"
	"time"

	"github.com/holiman/uint256"
)

type EventBlockNumbersErrors struct {
	ID          string
	EventName   string
	BlockNumber uint256.Int
	Options     []byte
	Errors      json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
