package block_validity_prover

import (
	"context"
	"time"
)

//go:generate mockgen -destination=mock_deposit_synchronizer.go -package=block_validity_prover -source=deposit_synchronizer.go

type DepositSynchronizer interface {
	Init(ctx context.Context) (err error)
	Start(tickerCurrentFile *time.Ticker) error
}
