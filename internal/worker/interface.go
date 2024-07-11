package worker

import (
	"context"
	"intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"os"
	"time"
)

type CurrentRootCountAndSiblings struct {
	TxTreeRoot tree.PoseidonHashOut
	Count      uint64
	Siblings   []*tree.PoseidonHashOut
}

type ReceiverWorker struct {
	Sender       string
	TransferHash string
	TransferData []*intMaxTypes.Transfer
}

type Worker interface {
	Init() (err error)
	Start(ctx context.Context, ticker *time.Ticker) error
	Receiver(input *ReceiverWorker) error
	CurrentDir() string
	CurrentFileName() string
	AvailableFiles() (list []*os.File)
	CurrentRootCountAndSiblingsFromRW(
		rw *ReceiverWorker,
	) (*CurrentRootCountAndSiblings, error)
}
