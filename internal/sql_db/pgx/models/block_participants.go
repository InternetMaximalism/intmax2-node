package models

import (
	"time"
)

type BlockParticipant struct {
	ID          string
	BlockNumber int64
	SenderId    string
	CreatedAt   time.Time
}
