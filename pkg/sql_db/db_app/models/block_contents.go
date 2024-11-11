package models

import (
	intMaxTypes "intmax2-node/internal/types"
	"time"

	"github.com/holiman/uint256"
)

type BlockContent struct {
	BlockContentID       string
	BlockNumber          uint32
	BlockHash            string
	PrevBlockHash        string
	DepositRoot          string
	DepositLeavesCounter uint32
	SignatureHash        string
	TxRoot               string
	AggregatedSignature  string
	AggregatedPublicKey  string
	MessagePoint         string
	Senders              []byte
	IsRegistrationBlock  bool
	BlockNumberL2        *uint256.Int
	BlockHashL2          string
	CreatedAt            time.Time
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
	BlockHash           string
	Senders             []intMaxTypes.ColumnSender
	DepositTreeRoot     string
	IsRegistrationBlock bool
}
