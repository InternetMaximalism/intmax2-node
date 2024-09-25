package models

const (
	DepositsAndAnalyzedReleyedEvent = "DepositsAndAnalyzedReleyed"
	WithdrawalSentMessageEvent      = "WithdrawalSentMessage"
	MessengerSentMessageEvent       = "MessengerSentMessage"
	WithdrawalsQueuedEvent          = "WithdrawalsQueued"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
