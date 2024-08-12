package models

import "time"

type BackupBalance struct {
	ID                    string    `json:"id"`
	UserAddress           string    `json:"user_address"`
	EncryptedBalanceProof string    `json:"encrypted_balance_proof"`
	EncryptedBalanceData  string    `json:"encrypted_balance_data"`
	EncryptedTxs          []string  `json:"encrypted_txs"`
	EncryptedTransfers    []string  `json:"encrypted_transfers"`
	EncryptedDeposits     []string  `json:"encrypted_deposits"`
	Signature             string    `json:"signature"`
	BlockNumber           uint64    `json:"block_number"`
	CreatedAt             time.Time `json:"created_at"`
}
