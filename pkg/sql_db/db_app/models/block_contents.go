package models

import (
	intMaxTypes "intmax2-node/internal/types"
	"time"
)

type BlockContent struct {
	BlockContentID      string
	BlockNumber         uint32
	BlockHash           string
	PrevBlockHash       string
	DepositRoot         string
	SignatureHash       string
	TxRoot              string
	AggregatedSignature string
	AggregatedPublicKey string
	MessagePoint        string
	Senders             []byte
	IsRegistrationBlock bool
	CreatedAt           time.Time
}

type BlockProof struct {
	BlockContentID string
	ValidityProof  []byte
}

type BlockContentWithProof struct {
	BlockContent
	ValidityProof []byte
}

type BlockHashAndSenders struct {
	BlockHash       string
	Senders         []intMaxTypes.ColumnSender
	DepositTreeRoot string
}
