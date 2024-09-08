package deposit_status_by_hash

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_deposit_status_by_hash.go -package=mocks -source=deposit_status_by_hash.go

type UCDepositStatusByHashInput struct {
	DepositHash string `json:"depositHash"`
}

type UCDepositStatusByHash struct {
	BlockNumber uint32 `json:"blockNumber"`
}

type UseCaseDepositStatusByHash interface {
	Do(ctx context.Context, input *UCDepositStatusByHashInput) (*UCDepositStatusByHash, error)
}
