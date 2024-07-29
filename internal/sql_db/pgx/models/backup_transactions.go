package models

import "time"

type BackupTransaction struct {
	ID          string    `json:"id"`
	Sender      string    `json:"sender"`
	EncryptedTx string    `json:"encrypted_tx"`
	BlockNumber uint64    `json:"block_number"`
	Signature   string    `json:"signature"`
	CreatedAt   time.Time `json:"created_at"`
}
