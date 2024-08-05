package messenger_relayer

import "context"

//go:generate mockgen -destination=../mocks/mock_messenger_relayer_mock.go -package=mocks -source=messenger_relayer_mock.go

// UseCaseMessengerRelayerMock describes MessengerRelayerMock contract.
type UseCaseMessengerRelayerMock interface {
	Do(ctx context.Context) error
}
