package deposit_analyzer

import "context"

//go:generate mockgen -destination=../mocks/mock_withdrawal_relayer.go -package=mocks -source=withdrawal_relayer.go

// UseCaseWithdrawal describes Withdrawal contract.
type UseCaseWithdrawalRelayer interface {
	Do(ctx context.Context) error
}
