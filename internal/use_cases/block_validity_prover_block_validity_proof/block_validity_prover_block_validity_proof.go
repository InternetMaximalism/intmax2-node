package block_validity_prover_block_validity_proof

import (
	"context"
	"intmax2-node/internal/block_validity_prover"
)

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_block_validity_proof.go -package=mocks -source=block_validity_prover_block_validity_proof.go

type UCBlockValidityProverBlockValidityProofInput struct {
	BlockNumber int64 `json:"blockNumber"`
}

type UCBlockValidityProverBlockValidityProof struct {
	ValidityPublicInputs *block_validity_prover.ValidityPublicInputs
	ValidityProof        *string
	Sender               []block_validity_prover.SenderLeaf
}

type UseCaseBlockValidityProverBlockValidityProof interface {
	Do(
		ctx context.Context,
		input *UCBlockValidityProverBlockValidityProofInput,
	) (*UCBlockValidityProverBlockValidityProof, error)
}
