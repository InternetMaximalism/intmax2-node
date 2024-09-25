package models

const (
	DepositsProcessedEvent = "DepositsProcessed"
	BlockPostedEvent       = "BlockPosted"
)

type EventBlockNumberForValidityProver struct {
	EventName                string
	LastProcessedBlockNumber uint64
}
