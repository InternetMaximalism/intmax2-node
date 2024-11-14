package models

import "time"

type BlockSender struct {
	ID        string
	Address   string
	PublicKey string
	CreatedAt time.Time
}
