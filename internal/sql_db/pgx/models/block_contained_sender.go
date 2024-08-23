package models

import (
	"time"
)

type BlockContainedSender struct {
	BlockContainedSenderID string
	BlockHash              string
	Sender                 string
	CreatedAt              time.Time
}
