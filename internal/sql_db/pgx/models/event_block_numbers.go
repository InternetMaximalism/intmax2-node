package models

const (
	DepositsAnalyzedEvent = "DepositsAnalyzed"
	DepositsRelayedEvent  = "DepositsRelayed"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
