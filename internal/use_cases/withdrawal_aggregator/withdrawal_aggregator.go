package withdrawal_aggregator

import "context"

//go:generate mockgen -destination=../mocks/mock_withdrawal_aggregator.go -package=mocks -source=withdrawal_aggregator.go

// UseCaseWithdrawalAggregator describes WithdrawalAggregator contract.
type UseCaseWithdrawalAggregator interface {
	Do(ctx context.Context) error
}
