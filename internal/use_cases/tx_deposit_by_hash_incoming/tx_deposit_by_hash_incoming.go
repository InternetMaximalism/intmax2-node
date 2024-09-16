package tx_deposit_by_hash_incoming

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=../mocks/mock_tx_deposit_by_hash_incoming.go -package=mocks -source=tx_deposit_by_hash_incoming.go

type UseCaseTxDepositByHashIncoming interface {
	Do(ctx context.Context, args []string, hash, userPrivateKey string) (json.RawMessage, error)
}
