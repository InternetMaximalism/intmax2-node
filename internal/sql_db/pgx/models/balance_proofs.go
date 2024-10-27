package models

import (
	"time"
)

type BalanceProof struct {
	ID                     string    `json:"id"`
	UserAddress            string    `json:"user_address"`
	BlockNumber            int64     `json:"block_number"` // uint32
	PrivateStateCommitment string    `json:"private_state_commitment"`
	BalanceProof           string    `json:"balance_proof"`
	CreatedAt              time.Time `json:"created_at"`
}
