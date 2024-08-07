package transaction

import (
	"intmax2-node/internal/worker"
)

//go:generate mockgen -destination=mock_worker_test.go -package=transaction_test -source=worker.go

type Worker interface {
	Receiver(input *worker.ReceiverWorker) error
}
