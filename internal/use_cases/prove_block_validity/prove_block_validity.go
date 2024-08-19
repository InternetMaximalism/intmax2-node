package prove_block_validity

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=../mocks/mock_prove_block_validity.go -package=mocks -source=prove_block_validity.go

const (
	SuccessMsg = "Withdrawal request accepted."
)

// Define the ProveBlockValidity struct
type UCProveBlockValidityInput struct {
	TransferHashes []string `json:"transfer_hashes"`
}

// UseCaseProveBlockValidity describes ProveBlockValidity
type UseCaseProveBlockValidity interface {
	Do(ctx context.Context, input *UCProveBlockValidityInput) (*[]mDBApp.Withdrawal, error)
}
