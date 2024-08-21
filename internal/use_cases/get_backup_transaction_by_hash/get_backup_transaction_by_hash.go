package get_backup_transaction_by_hash

import (
	"context"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_transaction_by_hash.go -package=mocks -source=get_backup_transaction_by_hash.go

const (
	NotFoundMessage = "Transaction hash not found."
)

type UCGetBackupTransactionByHashInput struct {
	Sender string `json:"sender"`
	TxHash string `json:"txHash"`
}

type UseCaseGetBackupTransactionByHash interface {
	Do(
		ctx context.Context, input *UCGetBackupTransactionByHashInput,
	) (*node.GetBackupTransactionByHashResponse_Data, error)
}
