package messenger_withdrawal_relayer

import "context"

//go:generate mockgen -destination=../mocks/mock_messenger_withdrawal_relayer.go -package=mocks -source=messenger_withdrawal_relayer.go

// UseCaseMessengerWithdrawalRelayer describes MessengerWithdrawalRelayer contract.
type UseCaseMessengerWithdrawalRelayer interface {
	Do(ctx context.Context) error
}
