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
	BlockNumber         sql.NullInt64
	AggregatedSignature string
	AggregatedPublicKey string
	Senders             []byte
	Status              sql.NullInt64
	CreatedAt           time.Time
	PostedAt            sql.NullTime
	SenderType          int64
	Options             []byte
}
