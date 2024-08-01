package block_signature

import (
	intMaxTree "intmax2-node/internal/tree"
	"intmax2-node/internal/worker"
)

//go:generate mockgen -destination=mock_worker_test.go -package=block_signature_test -source=worker.go

type Worker interface {
	SignTxTreeByAvailableFile(
		signature string,
		sf *worker.TransactionHashesWithSenderAndFile,
		txHash *intMaxTree.PoseidonHashOut,
		leafIndex uint64,
	) error
}
