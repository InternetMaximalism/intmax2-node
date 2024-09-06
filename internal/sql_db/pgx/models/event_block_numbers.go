package models

const (
	DepositsAndAnalyzedReleyedEvent = "DepositsAndAnalyzedReleyedEvent"
)

type EventBlockNumber struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
