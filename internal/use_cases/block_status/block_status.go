package block_status

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_block_status.go -package=mocks -source=block_status.go

type UCBlockStatusInput struct {
	TxTreeRoot string `json:"txTreeRoot"`
}

type UCBlockStatus struct {
	IsPosted    bool   `json:"isPosted"`
	BlockNumber uint32 `json:"blockNumber"`
}

// UseCaseBlockSignature describes BlockSignature contract.
type UseCaseBlockStatus interface {
	Do(ctx context.Context, input *UCBlockStatusInput) (*UCBlockStatus, error)
}
