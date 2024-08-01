package balance_checker

import "context"

//go:generate mockgen -destination=../mocks/mock_balance_checker.go -package=mocks -source=balance_checker.go

// UseCaseBalanceChecker describes BalanceChecker contract.
type UseCaseBalanceChecker interface {
	Do(ctx context.Context, args []string, userAddress string) error
}
