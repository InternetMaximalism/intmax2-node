package tx_resume_withdrawal

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_tx_resume_withdrawal.go -package=mocks -source=tx_resume_withdrawal.go

type UseCaseTxResumeWithdrawal interface {
	Do(ctx context.Context, recipientAddressHex string) error
}
