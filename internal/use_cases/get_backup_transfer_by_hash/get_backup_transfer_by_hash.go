package get_backup_transfer_by_hash

import (
	"context"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_transfer_by_hash.go -package=mocks -source=get_backup_transfer_by_hash.go

const (
	NotFoundMessage = "Transfer hash not found."
)

type UCGetBackupTransferByHashInput struct {
	Recipient    string `json:"recipient"`
	TransferHash string `json:"transferHash"`
}

type UseCaseGetBackupTransferByHash interface {
	Do(
		ctx context.Context, input *UCGetBackupTransferByHashInput,
	) (*node.GetBackupTransferByHashResponse_Data, error)
}
