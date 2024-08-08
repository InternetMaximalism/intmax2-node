package block_status

import "intmax2-node/internal/worker"

//go:generate mockgen -destination=mock_worker_test.go -package=block_status_test -source=worker.go

type Worker interface {
	// ExistsTxTreeRoot(txTreeRoot string) error
	TrHash(trHash string) (*worker.TransactionHashesWithSenderAndFile, error)
	TxTreeByAvailableFile(sf *worker.TransactionHashesWithSenderAndFile) (txTreeRoot *worker.TxTree, err error)
}
