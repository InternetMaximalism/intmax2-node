package models

import (
	"time"
)

const (
	TxPENDING = iota
	TxPROCESSING
	TxSUCCESS
	TxFAILED
)

type Transactions struct {
	TxID            string
	TxHash          string
	SenderPublicKey string
	SignatureID     string
	Status          int64
	CreatedAt       time.Time
}
