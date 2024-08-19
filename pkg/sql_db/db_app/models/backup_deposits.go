package models

import (
	"time"
)

type BackupDeposit struct {
	ID                string    `json:"id"`
	Recipient         string    `json:"recipient"`
	DepositDoubleHash string    `json:"deposit_double_hash"`
	EncryptedDeposit  string    `json:"encrypted_deposit"`
	BlockNumber       uint64    `json:"block_number"`
	CreatedAt         time.Time `json:"created_at"`
}
