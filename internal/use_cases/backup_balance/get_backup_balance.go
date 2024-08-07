package backup_balance

import (
	"context"
	"intmax2-node/internal/pb/gen/service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_balance.go -package=mocks -source=get_backup_balance.go

type UCGetBackupBalanceInput struct {
	Sender           string `json:"sender"`
	StartBlockNumber uint64 `json:"startBlockNumber"`
	Limit            uint64 `json:"limit"`
}

// UseCaseGetBackupBalance describes GetBackupBalance contract.
type UseCaseGetBackupBalance interface {
	Do(ctx context.Context, input *UCGetBackupBalanceInput) (*node.GetBackupBalanceResponse_Data, error)
}
