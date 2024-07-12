package models

import (
	"database/sql"
	"time"
)

type Signature struct {
	SignatureID     string
	Signature       string
	ProposalBlockID sql.NullString
	CreatedAt       time.Time
}
