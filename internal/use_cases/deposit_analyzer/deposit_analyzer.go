package deposit_analyzer

import "context"

//go:generate mockgen -destination=../mocks/mock_deposit_analyzer.go -package=mocks -source=deposit_analyzer.go

// UseCaseDepositAnalyzer describes DepositAnalyzer contract.
type UseCaseDepositAnalyzer interface {
	Do(ctx context.Context) error
}
