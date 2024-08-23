package models

import (
	"database/sql"
	"time"
)

type BackupDeposit struct {
	ID                string         `json:"id"`
	Recipient         string         `json:"recipient"`
	DepositDoubleHash sql.NullString `json:"deposit_double_hash"`
	EncryptedDeposit  string         `json:"encrypted_deposit"`
	BlockNumber       int64          `json:"block_number"`
	CreatedAt         time.Time      `json:"created_at"`
}

type ListOfBackupDeposit []BackupDeposit
