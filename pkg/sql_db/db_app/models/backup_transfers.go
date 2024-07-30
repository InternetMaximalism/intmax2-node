package models

import "time"

type BackupTransfer struct {
	ID                string    `json:"id"`
	EncryptedTransfer string    `json:"encrypted_transfer"`
	Recipient         string    `json:"recipient"`
	BlockNumber       uint64    `json:"block_number"`
	CreatedAt         time.Time `json:"created_at"`
}
