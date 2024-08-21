package post_backup_transaction

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_transaction.go -package=mocks -source=post_backup_transaction.go

const (
	SuccessMsg = "Backup transaction accepted."
)

type UCPostBackupTransactionInput struct {
	TxHash      string `json:"txHash"`
	EncryptedTx string `json:"encryptedTx"`
	Sender      string `json:"sender"`
	BlockNumber uint32 `json:"blockNumber"`
	Signature   string `json:"signature"`
}

// UseCasePostBackupTransaction describes PostBackupTransaction contract.
type UseCasePostBackupTransaction interface {
	Do(ctx context.Context, input *UCPostBackupTransactionInput) error
}
