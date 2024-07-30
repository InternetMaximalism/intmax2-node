package block_proposed

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/worker"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_block_proposed.go -package=mocks -source=block_proposed.go

type UCBlockProposed struct {
	TxRoot            string   `json:"txRoot"`
	TxTreeMerkleProof []string `json:"txTreeMerkleProof"`
	PublicKeys        []string `json:"publicKeys"`
}

type UCBlockProposedInput struct {
	Sender       string               `json:"sender"`
	DecodeSender *intMaxAcc.PublicKey `json:"-"`
	TxHash       string               `json:"txHash"`
	TxTree       *worker.TxTree       `json:"-"`
	Expiration   time.Time            `json:"expiration"`
	Signature    string               `json:"signature"`
}

// UseCaseBlockProposed describes BlockProposed contract.
type UseCaseBlockProposed interface {
	Do(ctx context.Context, input *UCBlockProposedInput) (*UCBlockProposed, error)
}
