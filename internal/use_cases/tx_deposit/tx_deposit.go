package tx_deposit

import (
	"context"
)

type UseCaseTxDeposit interface {
	Do(ctx context.Context, args []string, recipientAddressStr, amount, userPrivateKey string) error
}
