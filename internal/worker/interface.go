package worker

import (
	"context"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"os"
	"time"
)

type ComputeMerkleProof struct {
	Root     intMaxTree.PoseidonHashOut    `json:"root"`
	Siblings []*intMaxTree.PoseidonHashOut `json:"siblings"`
}

type CurrentRootCountAndSiblings struct {
	TransferTreeRoot intMaxTree.PoseidonHashOut    `json:"transferTreeRoot"`
	Count            uint64                        `json:"count"`
	Siblings         []*intMaxTree.PoseidonHashOut `json:"siblings"`
}

type ReceiverWorker struct {
	Sender        string
	Nonce         uint64
	TxHash        *intMaxTypes.Tx
	TransfersHash string
}

type SenderTxs map[string]*ReceiverWorker

type SenderInfo struct {
	Sender  *intMaxTypes.Sender
	TxsList map[string]*ReceiverWorker
}

type SenderTransfers struct {
	TxHash *intMaxTypes.PoseidonHashOut `json:"txHash"`
	// TxTreeLeafHash              *intMaxTree.PoseidonHashOut   `json:"txTreeLeafHash"`
	TxTreeRootHash              *intMaxTree.PoseidonHashOut   `json:"txTreeLeafHash"`
	TxTreeSiblings              []*intMaxTree.PoseidonHashOut `json:"txTreeSiblings"`
	CurrentRootCountAndSiblings *CurrentRootCountAndSiblings  `json:"currentRootCountAndSiblings"`
	ReceiverWorker              *ReceiverWorker               `json:"receiverWorker"`
}

type TxTree struct {
	RootHash *intMaxTree.PoseidonHashOut   `json:"rootHash"`
	Siblings []*intMaxTree.PoseidonHashOut `json:"siblings"`

	Sender          string                      `json:"sender"`
	TxTreeHash      *intMaxTree.PoseidonHashOut `json:"txTreeHash"`
	LeafIndexes     map[string]uint64           `json:"leafIndexes"`
	SenderTransfers []*SenderTransfers          `json:"senderTransfers"`
	Signature       string                      `json:"signature"`
}

type Worker interface {
	Init() (err error)
	Start(
		ctx context.Context,
		tickerCurrentFile, tickerSignaturesAvailableFiles *time.Ticker,
	) error
	Receiver(input *ReceiverWorker) error
	CurrentDir() string
	CurrentFileName() string
	AvailableFiles() (list []*os.File, err error)
	TrHash(trHash string) (*TransactionHashesWithSenderAndFile, error)
	TxTreeByAvailableFile(sf *TransactionHashesWithSenderAndFile) (txTreeRoot *TxTree, err error)
}
