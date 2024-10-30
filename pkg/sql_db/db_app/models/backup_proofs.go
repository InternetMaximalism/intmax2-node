package models

import (
	"time"
)

type BackupSenderProof struct {
	ID                         string    `json:"id"`
	EnoughBalanceProofBodyHash string    `json:"enough_balance_proof_body_hash"`
	LastBalanceProofBody       []byte    `json:"last_balance_proof_body"`
	BalanceTransitionProofBody []byte    `json:"balance_transition_proof_body"`
	CreatedAt                  time.Time `json:"created_at"`
}
