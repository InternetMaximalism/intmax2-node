package models

import (
	"time"
)

type Signature struct {
	SignatureID     string
	Signature       string
	ProposalBlockID string
	CreatedAt       time.Time
}
