package models

import "time"

type SpentProof struct {
	ID         string    `json:"id"`
	SpentProof string    `json:"spent_proof"`
	CreatedAt  time.Time `json:"created_at"`
}
