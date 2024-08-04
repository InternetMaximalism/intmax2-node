package models

const (
	DepositsAnalyzedEvent  = "DepositsAnalyzed"
	DepositsRelayedEvent   = "DepositsRelayed"
	SentMessageEvent       = "SentMessage"
	WithdrawalsQueuedEvent = "WithdrawalsQueued"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
