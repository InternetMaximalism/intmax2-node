package models

import (
	"database/sql"
	"time"
)

type Transactions struct {
	TxID            string
	TxHash          string
	SenderPublicKey string
	SignatureID     string
	Status          sql.NullInt64
	CreatedAt       time.Time
}
