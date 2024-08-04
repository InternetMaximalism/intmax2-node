package withdrawal_aggregator

import "context"

//go:generate mockgen -destination=../mocks/mock_withdrawal_relayer.go -package=mocks -source=withdrawal_relayer.go

// UseCaseWithdrawalRelayer describes WithdrawalRelayer contract.
type UseCaseWithdrawalRelayer interface {
	Do(ctx context.Context) error
}
