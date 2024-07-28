package models

import (
	"encoding/json"
	"time"
)

type Balance struct {
	ID          string
	UserAddress string
	TokenIndex  string
	Balance     string
	CreatedAt   time.Time
}

type BalanceBackup struct {
	ID                    string
	UserAddress           string
	BlockNumber           uint32
	EncryptedBalanceProof string
	EncryptedPublicInputs string
	EncryptedTxs          json.RawMessage
	EncryptedTransfers    json.RawMessage
	EncryptedDeposits     json.RawMessage
	CreatedAt             time.Time
}
