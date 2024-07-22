package backup_transaction

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_transaction.go -package=mocks -source=get_backup_transaction.go

type UCGetBackupTransactionContent struct {
	EncryptedTx string `json:"encryptedTx"`
	BlockNumber uint32 `json:"blockNumber"`
}

type UCGetBackupTransactionMeta struct {
	StartBlockNumber uint32 `json:"startBlockNumber"`
	EndBlockNumber   uint32 `json:"endBlockNumber"`
}

type UCGetBackupTransaction struct {
	Transactions []UCGetBackupTransactionContent `json:"transactions"`
	Meta         UCGetBackupTransactionMeta      `json:"meta"`
}

type UCGetBackupTransactionInput struct {
	Sender           string               `json:"user"`
	DecodeSender     *intMaxAcc.PublicKey `json:"-"`
	StartBlockNumber uint32               `json:"blockNumber"`
	Limit            uint                 `json:"limit"`
}

// UseCaseGetBackupTransaction describes GetBackupTransaction contract.
type UseCaseGetBackupTransaction interface {
	Do(ctx context.Context, input *UCGetBackupTransactionInput) (*UCGetBackupTransaction, error)
}
