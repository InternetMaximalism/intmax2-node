package block_validity_prover_deposits

import (
	"context"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_deposits.go -package=mocks -source=block_validity_prover_deposits.go

type UCBlockValidityProverDeposits struct {
	DepositId      uint32
	DepositHash    common.Hash
	DepositIndex   *uint32
	BlockNumber    *uint32
	IsSynchronized bool
	DepositLeaf    *intMaxTree.DepositLeaf
	Sender         string
}

type UCBlockValidityProverDepositsInput struct {
	DepositHashes        []string      `json:"depositHashes"`
	ConvertDepositHashes []common.Hash `json:"-"`
}

type UseCaseBlockValidityProverDeposits interface {
	Do(
		ctx context.Context,
		input *UCBlockValidityProverDepositsInput,
	) ([]*UCBlockValidityProverDeposits, error)
}
