package backup_transaction

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_transaction.go -package=mocks -source=post_backup_transaction.go

type UCPostBackupTransaction struct {
	Message string `json:"message"`
}

type UCPostBackupTransactionInput struct {
	Sender       string               `json:"user"`
	DecodeSender *intMaxAcc.PublicKey `json:"-"`
	BlockNumber  uint32               `json:"blockNumber"`
	EncryptedTx  string               `json:"encryptedTx"`
	Signature    string               `json:"signature"`
}

// UseCasePostBackupTransaction describes PostBackupTransaction contract.
type UseCasePostBackupTransaction interface {
	Do(ctx context.Context, input *UCPostBackupTransactionInput) (*UCPostBackupTransaction, error)
}
