package tx_transfer

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_tx_transfer.go -package=mocks -source=tx_transfer.go

type UseCaseTxTransfer interface {
	Do(ctx context.Context, args []string, recipientAddressHex, amount, userPrivateKey string) error
}
