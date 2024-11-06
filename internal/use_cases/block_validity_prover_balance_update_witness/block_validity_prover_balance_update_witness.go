package block_validity_prover_balance_update_witness

import (
	"context"
	intMaxTree "intmax2-node/internal/tree"
)

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_balance_update_witness.go -package=mocks -source=block_validity_prover_balance_update_witness.go

type UCBlockValidityProverBalanceUpdateWitnessInput struct {
	User                                          string `json:"user"`
	CurrentBlockNumber                            int64  `json:"currentBlockNumber"`
	IsInvalidCurrentBlockNumber                   bool   `json:"-"`
	TargetBlockNumber                             int64  `json:"targetBlockNumber"`
	IsInvalidTargetBlockNumber                    bool   `json:"-"`
	IsTargetBlockNumberMoreThenCurrentBlockNumber bool   `json:"-"`
	IsCurrentBlockNumberLessThenTargetBlockNumber bool   `json:"-"`
	IsPrevAccountTree                             bool   `json:"isPrevAccountTree"`
}

type UCBlockValidityProverBalanceUpdateWitness struct {
	IsPrevAccountTree      bool
	ValidityProof          string
	BlockMerkleProof       intMaxTree.BlockHashMerkleProof
	AccountMembershipProof *intMaxTree.IndexedMembershipProof
}

type UseCaseBlockValidityProverBalanceUpdateWitness interface {
	Do(
		ctx context.Context,
		input *UCBlockValidityProverBalanceUpdateWitnessInput,
	) (*UCBlockValidityProverBalanceUpdateWitness, error)
}
