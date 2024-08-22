package server

import (
	"context"
	"time"
)

type DepositSynchronizer interface {
	Init(ctx context.Context) (err error)
	Start(ctx context.Context, tickerCurrentFile *time.Ticker) error
}
