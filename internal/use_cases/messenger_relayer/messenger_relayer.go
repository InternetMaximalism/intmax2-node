package messenger_relayer

import "context"

//go:generate mockgen -destination=../mocks/mock_messenger_relayer.go -package=mocks -source=messenger_relayer.go

// UseCaseRelayMessenger describes MessengerRelayer contract.
type UseCaseMessengerRelayer interface {
	Do(ctx context.Context) error
}
