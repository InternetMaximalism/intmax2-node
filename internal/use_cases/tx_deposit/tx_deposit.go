package tx_deposit

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_tx_deposit.go -package=mocks -source=tx_deposit.go

type UseCaseTxDeposit interface {
	Do(ctx context.Context, args []string, recipientAddressHex, amount, userPrivateKey string) error
}
