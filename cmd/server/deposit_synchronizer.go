package server

import (
	"context"
	"time"
)

type DepositSynchronizer interface {
	Init() (err error)
	Start(ctx context.Context, tickerCurrentFile *time.Ticker) error
}
