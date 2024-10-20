package models

import (
	"encoding/json"
	"time"

	"github.com/holiman/uint256"
)

type L2BatchIndex struct {
	L2BatchIndex          uint256.Int
	Options               json.RawMessage
	L1VerifiedBatchTxHash string
	CreatedAt             time.Time
}
