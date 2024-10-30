package server

import (
	"context"
	"time"
)

//go:generate mockgen -destination=mock_deposit_synchronizer.go -package=server -source=deposit_synchronizer.go

type DepositSynchronizer interface {
	Init(ctx context.Context) (err error)
	Start(tickerCurrentFile *time.Ticker) error
}
