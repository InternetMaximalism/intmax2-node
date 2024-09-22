package deposit_synchronizer

import (
	"context"
	"time"
)

type DepositSynchronizer interface {
	Init(ctx context.Context) error
	Start(tickerEventWatcher *time.Ticker) (err error)
}
