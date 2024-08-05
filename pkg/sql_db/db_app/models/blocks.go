package models

import (
	"time"
)

const (
	ST_PUBLIC_KEY = iota
	ST_ACCOUNT_ID
)

const (
	B_PENDING = iota
	B_PROCESSING
	B_SUCCESS
	B_FAILED
)

type Block struct {
	ProposalBlockID     string
	BuilderPublicKey    string
	TxRoot              string
	BlockHash           string
	BlockNumber         *int64
	AggregatedSignature string
	AggregatedPublicKey string
	Senders             []byte
	Status              *int64
	CreatedAt           time.Time
	PostedAt            *time.Time
	SenderType          int64
	Options             []byte
}
