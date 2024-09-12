package models

const (
	DepositsProcessedEvent = "DepositsProcessed"
)

type EventBlockNumberForValidityProver struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
