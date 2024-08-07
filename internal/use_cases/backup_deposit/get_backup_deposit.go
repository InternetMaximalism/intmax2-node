package backup_deposit

import (
	"context"
	"intmax2-node/internal/pb/gen/service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_deposit.go -package=mocks -source=get_backup_deposit.go

type UCGetBackupDepositInput struct {
	Sender           string `json:"sender"`
	StartBlockNumber uint64 `json:"startBlockNumber"`
	Limit            uint64 `json:"limit"`
}

// UseCaseGetBackupDeposit describes GetBackupDeposit contract.
type UseCaseGetBackupDeposit interface {
	Do(ctx context.Context, input *UCGetBackupDepositInput) (*node.GetBackupDepositResponse_Data, error)
}
