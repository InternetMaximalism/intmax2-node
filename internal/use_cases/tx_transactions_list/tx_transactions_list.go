package tx_transactions_list

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=../mocks/mock_tx_transactions_list.go -package=mocks -source=tx_transactions_list.go

type UseCaseTxTransactionsList interface {
	Do(ctx context.Context, args []string, startBlockNumber, limit, userPrivateKey string) (json.RawMessage, error)
}
