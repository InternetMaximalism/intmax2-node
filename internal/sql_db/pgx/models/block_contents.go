package models

import (
	"time"
)

type BlockContent struct {
	BlockContentID      string
	BlockNumber         int64
	BlockHash           string
	PrevBlockHash       string
	DepositRoot         string
	IsRegistrationBlock bool
	TxRoot              string
	AggregatedSignature string
	AggregatedPublicKey string
	MessagePoint        string
	Senders             []byte
	CreatedAt           time.Time
}
