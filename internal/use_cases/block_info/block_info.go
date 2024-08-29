package block_info

import "context"

//go:generate mockgen -destination=../mocks/mock_block_info.go -package=mocks -source=block_info.go

type UCBlockInfo struct {
	ScrollAddress string            `json:"scrollAddress"`
	IntMaxAddress string            `json:"intMaxAddress"`
	TransferFee   map[string]string `json:"transferFee"`
	Difficulty    int64             `json:"difficulty"`
}

// UseCaseBlockInfo describes BlockInfo contract.
type UseCaseBlockInfo interface {
	Do(ctx context.Context) (*UCBlockInfo, error)
}
