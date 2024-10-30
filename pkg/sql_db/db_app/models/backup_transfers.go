package models

import (
	mFL "intmax2-node/internal/sql_filter/models"
	"math/big"
	"time"
)

type BackupTransfer struct {
	ID                 string    `json:"id"`
	TransferDoubleHash string    `json:"transfer_double_hash"`
	EncryptedTransfer  string    `json:"encrypted_transfer"`
	Recipient          string    `json:"recipient"`
	BlockNumber        uint64    `json:"block_number"`
	CreatedAt          time.Time `json:"created_at"`
}

type ListOfBackupTransfer []BackupTransfer

type PaginationOfListOfBackupTransfersInput struct {
	Direction mFL.Direction
	Offset    int
	Cursor    *CursorBaseOfListOfBackupTransfers
}

type CursorBaseOfListOfBackupTransfers struct {
	BN           *big.Int
	SortingValue *big.Int
}

type PaginationOfListOfBackupTransfers struct {
	Offset int
	Cursor *CursorListOfBackupTransfers
}

type CursorListOfBackupTransfers struct {
	Prev *CursorBaseOfListOfBackupTransfers
	Next *CursorBaseOfListOfBackupTransfers
}
