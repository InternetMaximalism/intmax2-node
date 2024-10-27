package models

import (
	"time"
)

type BlockContainedSender struct {
	BlockContainedSenderID string
	BlockNumber            uint32
	SenderId               string
	CreatedAt              time.Time
}
