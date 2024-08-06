package server

import (
	"context"
	intMaxTree "intmax2-node/internal/tree"
	"intmax2-node/internal/worker"
	"time"
)

type Worker interface {
	Init() (err error)
	Start(
		ctx context.Context,
		tickerCurrentFile, tickerSignaturesAvailableFiles *time.Ticker,
	) error
	Receiver(input *worker.ReceiverWorker) error
	TrHash(trHash string) (*worker.TransactionHashesWithSenderAndFile, error)
	TxTreeByAvailableFile(sf *worker.TransactionHashesWithSenderAndFile) (txTreeRoot *worker.TxTree, err error)
	SignTxTreeByAvailableFile(
		signature string,
		sf *worker.TransactionHashesWithSenderAndFile,
		txHash *intMaxTree.PoseidonHashOut,
		leafIndex uint64,
	) error
	ExistsTxTreeRoot(txTreeRoot string) error
}
