package models

import (
	mFL "intmax2-node/internal/sql_filter/models"
	"math/big"
	"time"
)

type BackupDeposit struct {
	ID                string    `json:"id"`
	Recipient         string    `json:"recipient"`
	DepositDoubleHash string    `json:"deposit_double_hash"`
	EncryptedDeposit  string    `json:"encrypted_deposit"`
	BlockNumber       int64     `json:"block_number"`
	CreatedAt         time.Time `json:"created_at"`
}

type ListOfBackupDeposit []BackupDeposit

type PaginationOfListOfBackupDepositsInput struct {
	Direction mFL.Direction
	Offset    int
	Cursor    *CursorBaseOfListOfBackupDeposits
}

type CursorBaseOfListOfBackupDeposits struct {
	BN           *big.Int
	SortingValue *big.Int
}

type PaginationOfListOfBackupDeposits struct {
	Offset int
	Cursor *CursorListOfBackupDeposits
}

type CursorListOfBackupDeposits struct {
	Prev *CursorBaseOfListOfBackupDeposits
	Next *CursorBaseOfListOfBackupDeposits
}
