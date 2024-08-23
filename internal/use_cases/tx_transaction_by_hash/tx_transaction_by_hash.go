package tx_transaction_by_hash

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=../mocks/mock_tx_transaction_by_hash.go -package=mocks -source=tx_transaction_by_hash.go

type UseCaseTxTransactionByHash interface {
	Do(ctx context.Context, args []string, hash, userPrivateKey string) (json.RawMessage, error)
}
