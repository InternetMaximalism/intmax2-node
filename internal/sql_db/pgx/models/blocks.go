package models

import (
	"database/sql"
	"time"
)

type Block struct {
	ProposalBlockID     string
	BuilderPublicKey    string
	TxRoot              string
	BlockHash           sql.NullString
	AggregatedSignature string
	AggregatedPublicKey string
	Status              sql.NullInt64
	CreatedAt           time.Time
	PostedAt            sql.NullTime
	SenderType          int64
	Options             []byte
}
