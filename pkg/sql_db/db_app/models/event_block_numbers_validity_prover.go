package models

const (
	DepositsProcessedEvent = "DepositsProcessedEvent"
)

type EventBlockNumberForValidityProver struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
