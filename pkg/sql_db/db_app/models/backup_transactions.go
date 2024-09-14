package models

import (
	mFL "intmax2-node/internal/sql_filter/models"
	"math/big"
	"time"
)

type BackupTransaction struct {
	ID              string    `json:"id"`
	Sender          string    `json:"sender"`
	TxDoubleHash    string    `json:"tx_double_hash"`
	EncryptedTx     string    `json:"encrypted_tx"`
	EncodingVersion int64     `json:"encoding_version"`
	BlockNumber     int64     `json:"block_number"`
	Signature       string    `json:"signature"`
	CreatedAt       time.Time `json:"created_at"`
}

type ListOfBackupTransaction []BackupTransaction

type PaginationOfListOfBackupTransactionsInput struct {
	Direction mFL.Direction
	Offset    int
	Cursor    *CursorBaseOfListOfBackupTransactions
}

type CursorBaseOfListOfBackupTransactions struct {
	BN           *big.Int
	SortingValue *big.Int
}

type PaginationOfListOfBackupTransactions struct {
	Offset int
	Cursor *CursorListOfBackupTransactions
}

type CursorListOfBackupTransactions struct {
	Prev *CursorBaseOfListOfBackupTransactions
	Next *CursorBaseOfListOfBackupTransactions
}
