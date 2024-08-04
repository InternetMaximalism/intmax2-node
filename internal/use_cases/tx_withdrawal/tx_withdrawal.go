package tx_withdrawal

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_tx_withdrawal.go -package=mocks -source=tx_withdrawal.go

type UseCaseTxWithdrawal interface {
	Do(ctx context.Context, args []string, recipientAddressHex, amount, userPrivateKey string) error
}
