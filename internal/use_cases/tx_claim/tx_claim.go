package tx_withdrawal

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_tx_claim.go -package=mocks -source=tx_claim.go

type UseCaseTxClaim interface {
	Do(ctx context.Context, args []string, recipientEthAddress string) error
}
