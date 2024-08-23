package deposit_synchronizer

import (
	"context"
	"time"
)

type DepositSynchronizer interface {
	Init(ctx context.Context) error
	Start(
		ctx context.Context,
		tickerEventWatcher *time.Ticker,
	) (err error)
}
