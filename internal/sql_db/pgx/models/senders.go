package models

import "time"

type Sender struct {
	ID        string
	Address   string
	PublicKey string
	CreatedAt time.Time
}
