package models

const (
	DepositedEvent                  = "Deposited"
	DepositsAndAnalyzedRelayedEvent = "DepositsAndAnalyzedReleyed"
	WithdrawalSentMessageEvent      = "WithdrawalSentMessage"
	MessengerSentMessageEvent       = "MessengerSentMessage"
	WithdrawalsQueuedEvent          = "WithdrawalsQueued"
	BlockPostedEvent                = "BlockPosted"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
