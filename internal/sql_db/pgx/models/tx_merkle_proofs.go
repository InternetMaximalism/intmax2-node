package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/holiman/uint256"
)

type TxMerkleProofs struct {
	ID              string
	SenderPublicKey string
	SignatureID     sql.NullString
	TxHash          string
	TxTreeIndex     *uint256.Int
	TxMerkleProof   json.RawMessage
	TxTreeRoot      string
	ProposalBlockID string
	CreatedAt       time.Time
}
