package block_validity_prover_account

import (
	"context"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_account.go -package=mocks -source=block_validity_prover_account.go

type UCBlockValidityProverAccountInput struct {
	Address string `json:"address"`
}

type UCBlockValidityProverAccount struct {
	AccountID *uint256.Int
}

type UseCaseBlockValidityProverAccount interface {
	Do(ctx context.Context, input *UCBlockValidityProverAccountInput) (*UCBlockValidityProverAccount, error)
}
