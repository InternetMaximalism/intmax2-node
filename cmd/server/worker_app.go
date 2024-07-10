package server

import (
	"context"
	"intmax2-node/internal/worker"
	"time"
)

type Worker interface {
	Init() (err error)
	Start(ctx context.Context, ticker *time.Ticker) error
	Receiver(input *worker.ReceiverWorker) error
}
