package models

const (
	DepositsAnalyzedEvent      = "DepositsAnalyzed"
	DepositsRelayedEvent       = "DepositsRelayed"
	WithdrawalSentMessageEvent = "WithdrawalSentMessage"
	MessengerSentMessageEvent  = "MessengerSentMessage"
	WithdrawalsQueuedEvent     = "WithdrawalsQueued"
	BlockPostedEvent           = "BlockPosted"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
