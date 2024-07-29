package models

const (
	DepositsAnalyzedEvent = "DepositsAnalyzed"
	DepositsRelayedEvent  = "DepositsRelayed"
	SentMessageEvent      = "SentMessageEvent"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber int64
}
