package withdrawal_relayer

import "context"

//go:generate mockgen -destination=../mocks/mock_withdrawal_relayer_mock.go -package=mocks -source=withdrawal_relayer_mock.go

// UseCaseWithdrawalRelayer describes WithdrawalRelayer contract.
type UseCaseWithdrawalRelayerMock interface {
	Do(ctx context.Context) error
}
