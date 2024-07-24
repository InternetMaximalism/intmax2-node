package models

import (
	"time"
)

type Balance struct {
	ID          string
	UserAddress string
	TokenIndex  string
	Balance     string
	CreatedAt   time.Time
}
