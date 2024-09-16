package get_backup_deposit_by_hash

import (
	"context"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_deposit_by_hash.go -package=mocks -source=get_backup_deposit_by_hash.go

const (
	NotFoundMessage = "Deposit hash not found."
)

type UCGetBackupDepositByHashInput struct {
	Recipient   string `json:"recipient"`
	DepositHash string `json:"depositHash"`
}

type UseCaseGetBackupDepositByHash interface {
	Do(
		ctx context.Context, input *UCGetBackupDepositByHashInput,
	) (*node.GetBackupDepositByHashResponse_Data, error)
}
