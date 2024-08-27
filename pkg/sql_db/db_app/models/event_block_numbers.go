package models

const (
	DepositsAndAnalyzedReleyedEvent = "DepositsAndAnalyzedReleyed"
	WithdrawalSentMessageEvent      = "WithdrawalSentMessage"
	MessengerSentMessageEvent       = "MessengerSentMessage"
	WithdrawalsQueuedEvent          = "WithdrawalsQueued"
	BlockPostedEvent                = "BlockPosted"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
