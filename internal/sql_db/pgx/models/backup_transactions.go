package models

import (
	"database/sql"
	"time"
)

type BackupTransaction struct {
	ID              string         `json:"id"`
	Sender          string         `json:"sender"`
	TxDoubleHash    sql.NullString `json:"tx_double_hash"`
	EncryptedTx     string         `json:"encrypted_tx"`
	EncodingVersion int64          `json:"encoding_version"`
	BlockNumber     int64          `json:"block_number"`
	Signature       string         `json:"signature"`
	CreatedAt       time.Time      `json:"created_at"`
}

type ListOfBackupTransaction []BackupTransaction
