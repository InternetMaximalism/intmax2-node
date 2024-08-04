package models

const (
	DepositsAnalyzedEvent  = "DepositsAnalyzed"
	DepositsRelayedEvent   = "DepositsRelayed"
	SentMessageEvent       = "SentMessage"
	WithdrawalsQueuedEvent = "WithdrawalsQueued"
	BlockPostedEvent       = "BlockPosted"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
