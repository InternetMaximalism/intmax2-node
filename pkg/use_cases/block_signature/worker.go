package block_signature

import (
	"intmax2-node/internal/worker"
)

//go:generate mockgen -destination=mock_worker_test.go -package=block_signature_test -source=worker.go

type Worker interface {
	SignTxTreeByAvailableFile(signature string, sf *worker.TransactionHashesWithSenderAndFile) error
}
