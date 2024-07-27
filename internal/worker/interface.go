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
	Sender       string                  `json:"sender"`
	Nonce        uint64                  `json:"nonce"`
	TxHash       *intMaxTypes.Tx         `json:"txHash"`
	TransferHash string                  `json:"transferHash"`
	TransferData []*intMaxTypes.Transfer `json:"transferData"`
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
	Sender         string                        `json:"sender"`
	TxHash         *intMaxTree.PoseidonHashOut   `json:"txTreeHash"`
	LeafIndex      uint64                        `json:"leafIndex"`
	TxTreeSiblings []*intMaxTree.PoseidonHashOut `json:"txTreeSiblings"`
	TxTreeRootHash *intMaxTree.PoseidonHashOut   `json:"txTreeRootHash"`
	// SenderTransfers []*SenderTransfers            `json:"senderTransfers"`
	Signature string `json:"signature"`
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
	AvailableFiles() (list []*os.File)
	TrHash(trHash string) (*TransactionHashesWithSenderAndFile, error)
	TxTreeByAvailableFile(sf *TransactionHashesWithSenderAndFile) (txTreeRoot *TxTree, err error)
}
