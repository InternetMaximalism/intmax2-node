package backup_transaction

import (
	"context"
	"intmax2-node/internal/pb/gen/service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_transaction.go -package=mocks -source=get_backup_transaction.go

type UCGetBackupTransactionInput struct {
	Sender           string `json:"sender"`
	StartBlockNumber uint64 `json:"startBlockNumber"`
	Limit            uint64 `json:"limit"`
}

// UseCaseGetBackupTransaction describes GetBackupTransaction contract.
type UseCaseGetBackupTransaction interface {
	Do(ctx context.Context, input *UCGetBackupTransactionInput) (*node.GetBackupTransactionResponse_Data, error)
}
