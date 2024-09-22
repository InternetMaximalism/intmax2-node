package server

import (
	"context"
	"time"
)

type DepositSynchronizer interface {
	Init(ctx context.Context) (err error)
	Start(tickerCurrentFile *time.Ticker) error
}
