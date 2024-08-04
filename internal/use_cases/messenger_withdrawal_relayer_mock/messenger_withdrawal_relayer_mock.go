package messenger_withdrawal_relayer_mock

import "context"

//go:generate mockgen -destination=../mocks/mock_messenger_withdrawal_relayer_mock.go -package=mocks -source=messenger_withdrawal_relayer_mock.go

// UseCaseMessengerWithdrawalRelayer describes MessengerWithdrawalRelayer contract.
type UseCaseMessengerWithdrawalRelayerMock interface {
	Do(ctx context.Context) error
}
