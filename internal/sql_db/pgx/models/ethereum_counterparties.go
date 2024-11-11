package models

import "time"

type EthereumCounterparty struct {
	ID        string
	Address   string
	CreatedAt time.Time
}
