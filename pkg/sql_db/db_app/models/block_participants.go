package models

import (
	"time"
)

type BlockParticipant struct {
	ID          string
	BlockNumber uint32
	SenderId    string
	CreatedAt   time.Time
}
