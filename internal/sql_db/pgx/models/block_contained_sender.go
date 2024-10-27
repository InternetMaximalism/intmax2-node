package models

import (
	"time"
)

type BlockContainedSender struct {
	BlockContainedSenderID string
	BlockNumber            int64
	SenderId               string
	CreatedAt              time.Time
}
