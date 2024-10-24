package get_intmax_block_info

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=../mocks/mock_get_intmax_block_info.go -package=mocks -source=get_intmax_block_info.go

type UseCaseGetINTMAXBlockInfo interface {
	Do(ctx context.Context, args []string, hash string, number uint64) (json.RawMessage, error)
}
