package models

import (
	"database/sql"
	"time"
)

type BackupTransfer struct {
	ID                               string         `json:"id"`
	TransferDoubleHash               sql.NullString `json:"transfer_double_hash"`
	EncryptedTransfer                string         `json:"encrypted_transfer"`
	SenderLastBalanceProofBody       []byte         `json:"sender_last_balance_proof_body"`
	SenderBalanceTransitionProofBody []byte         `json:"sender_balance_transition_proof_body"`
	Recipient                        string         `json:"recipient"`
	BlockNumber                      uint64         `json:"block_number"`
	CreatedAt                        time.Time      `json:"created_at"`
}

type ListOfBackupTransfer []BackupTransfer
