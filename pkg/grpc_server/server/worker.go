package server

import (
	"intmax2-node/internal/worker"
)

//go:generate mockgen -destination=mock_worker_test.go -package=server_test -source=worker.go

type Worker interface {
	Receiver(input *worker.ReceiverWorker) error
	TrHash(trHash string) (*worker.TransferHashesWithSenderAndFile, error)
	TxTreeByAvailableFile(sf *worker.TransferHashesWithSenderAndFile) (txTreeRoot *worker.TxTree, err error)
}
