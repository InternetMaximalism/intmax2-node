package backup_transfer

import (
	"context"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_transfers.go -package=mocks -source=get_backup_transfers.go

type UCGetBackupTransferInput struct {
	Sender           string `json:"sender"`
	StartBlockNumber uint64 `json:"startBlockNumber"`
	Limit            uint64 `json:"limit"`
}

// UseCaseGetBackupTransfers describes GetBackupTransfers contract.
type UseCaseGetBackupTransfers interface {
	Do(ctx context.Context, input *UCGetBackupTransferInput) (*node.GetBackupTransfersResponse_Data, error)
}
