package block_validity_prover_info

import "context"

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_info.go -package=mocks -source=block_validity_prover_info.go

type UCBlockValidityProverInfo struct {
	DepositIndex int64 `json:"depositIndex"`
	BlockNumber  int64 `json:"blockNumber"`
}

type UseCaseBlockValidityProverInfo interface {
	Do(ctx context.Context) (*UCBlockValidityProverInfo, error)
}
