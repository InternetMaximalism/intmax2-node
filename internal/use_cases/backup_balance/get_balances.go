package backup_balance

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_balance.go -package=mocks -source=get_backup_balance.go

type TokenBalance struct {
	TokenIndex int    `json:"tokenIndex"`
	Amount     string `json:"amount"`
}

type UCGetBalances struct {
	Balances []*TokenBalance `json:"balances"`
}

type UCGetBalancesInput struct {
	Address string `json:"address"`
}

// UseCaseGetBalances describes GetBalances contract.
type UseCaseGetBalances interface {
	Do(ctx context.Context, input *UCGetBalancesInput) (*UCGetBalances, error)
}
