package server

import (
	"intmax2-node/internal/worker"
)

//go:generate mockgen -destination=mock_worker_test.go -package=server_test -source=worker.go

type Worker interface {
	Receiver(input *worker.ReceiverWorker) error
	TrHash(trHash string) (*worker.TransactionHashesWithSenderAndFile, error)
	TxTreeByAvailableFile(sf *worker.TransactionHashesWithSenderAndFile) (txTreeRoot *worker.TxTree, err error)
	SignTxTreeByAvailableFile(signature string, sf *worker.TransactionHashesWithSenderAndFile) error
}
