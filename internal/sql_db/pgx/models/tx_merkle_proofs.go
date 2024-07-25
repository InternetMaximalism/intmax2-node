package models

import (
	"encoding/json"
	"time"

	"github.com/holiman/uint256"
)

type TxMerkleProofs struct {
	ID              string
	SenderPublicKey string
	TxHash          string
	TxID            string
	TxTreeIndex     *uint256.Int
	TxMerkleProof   json.RawMessage
	CreatedAt       time.Time
}
