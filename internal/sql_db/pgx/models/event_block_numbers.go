package models

import (
	"time"
)

const (
	DepositsAnalyzed = "DepositsAnalyzed"
)

type EventBlockNumber struct {
	ID                       string
	EventName                string
	LastProcessedBlockNumber int64
	CreatedAt                time.Time
}
