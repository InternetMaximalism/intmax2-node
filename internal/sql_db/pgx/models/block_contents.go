package models

import (
	"database/sql"
	"time"

	"github.com/holiman/uint256"
)

type BlockContent struct {
	BlockContentID      string
	BlockNumber         int64
	BlockHash           string
	PrevBlockHash       string
	DepositRoot         string
	SignatureHash       string
	IsRegistrationBlock bool
	TxRoot              string
	AggregatedSignature string
	AggregatedPublicKey string
	MessagePoint        string
	Senders             []byte
	ValidityProof       []byte
	BlockNumberL2       *uint256.Int
	BlockHashL2         sql.NullString
	CreatedAt           time.Time
}

type BlockProof struct {
	BlockContentID string
	ValidityProof  []byte
}
