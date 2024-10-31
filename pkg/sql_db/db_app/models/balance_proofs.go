package models

import "time"

type BalanceProof struct {
	ID                     string
	UserStateID            string
	UserAddress            string
	BlockNumber            int64
	PrivateStateCommitment string
	BalanceProof           []byte
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
