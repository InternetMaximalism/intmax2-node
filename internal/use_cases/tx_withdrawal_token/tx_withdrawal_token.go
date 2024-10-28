package tx_withdrawal_token

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_tx_withdrawal_token.go -package=mocks -source=tx_withdrawal_token.go

type UseCaseTxWithdrawalToken interface {
	Do(
		ctx context.Context,
		args []string,
		recipientAddressHex, amount, userPrivateKey string,
		resumeIncompleteWithdrawals bool,
	) error
}
