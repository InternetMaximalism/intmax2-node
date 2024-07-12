package server

import (
	"context"
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
	TrHash(trHash string) (*worker.TransferHashesWithSenderAndFile, error)
	TxTreeByAvailableFile(sf *worker.TransferHashesWithSenderAndFile) (txTreeRoot *worker.TxTree, err error)
}
