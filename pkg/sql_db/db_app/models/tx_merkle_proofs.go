package models

import (
	"encoding/json"
	"time"

	"github.com/holiman/uint256"
)

type TxMerkleProofs struct {
	ID              string
	SenderPublicKey string
	SignatureID     string
	TxHash          string
	TxTreeIndex     *uint256.Int
	TxMerkleProof   json.RawMessage
	TxTreeRoot      string
	ProposalBlockID string
	CreatedAt       time.Time
}
