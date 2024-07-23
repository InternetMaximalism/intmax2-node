package deposit_relayer

import "context"

//go:generate mockgen -destination=../mocks/mock_deposit_relayer.go -package=mocks -source=deposit_relayer.go

// UseCaseDepositRelayer describes DepositRelayer contract.
type UseCaseDepositRelayer interface {
	Do(ctx context.Context) error
}
