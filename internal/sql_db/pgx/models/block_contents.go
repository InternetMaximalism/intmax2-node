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
	SignatureHash       string
	IsRegistrationBlock bool
	TxRoot              string
	AggregatedSignature string
	AggregatedPublicKey string
	MessagePoint        string
	Senders             []byte
	ValidityProof       []byte
	CreatedAt           time.Time
}
