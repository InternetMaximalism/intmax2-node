package backup_balance_proof

import (
	"context"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_balance_proofs.go -package=mocks -source=get_backup_balance_proofs.go

type UCGetBackupBalanceProofsInput struct {
	Hashes []string `json:"hashes"`
}

// UseCaseGetBackupBalances describes GetBackupBalances contract.
type UseCaseGetBackupBalanceProofs interface {
	Do(ctx context.Context, input *UCGetBackupBalanceProofsInput) (*node.GetBackupBalanceProofsResponse_Data, error)
}
