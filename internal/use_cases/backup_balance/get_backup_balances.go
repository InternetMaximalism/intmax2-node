package backup_balance

import (
	"context"
	"intmax2-node/internal/pb/gen/service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_balances.go -package=mocks -source=get_backup_balances.go

type UCGetBackupBalancesInput struct {
	Sender           string `json:"sender"`
	StartBlockNumber uint64 `json:"startBlockNumber"`
	Limit            uint64 `json:"limit"`
}

// UseCaseGetBackupBalances describes GetBackupBalances contract.
type UseCaseGetBackupBalances interface {
	Do(ctx context.Context, input *UCGetBackupBalancesInput) (*node.GetBackupBalancesResponse_Data, error)
}
