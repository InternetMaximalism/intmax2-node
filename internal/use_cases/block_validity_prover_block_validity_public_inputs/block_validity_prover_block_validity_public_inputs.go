package block_validity_prover_block_validity_public_inputs

import (
	"context"
	"intmax2-node/internal/block_validity_prover"
)

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_block_validity_public_inputs.go -package=mocks -source=block_validity_prover_block_validity_public_inputs.go

type UCBlockValidityProverBlockValidityPublicInputsInput struct {
	BlockNumber int64 `json:"blockNumber"`
}

type UCBlockValidityProverBlockValidityPublicInputs struct {
	ValidityPublicInputs *block_validity_prover.ValidityPublicInputs
	Sender               []block_validity_prover.SenderLeaf
}

type UseCaseBlockValidityProverBlockValidityPublicInputs interface {
	Do(
		ctx context.Context,
		input *UCBlockValidityProverBlockValidityPublicInputsInput,
	) (*UCBlockValidityProverBlockValidityPublicInputs, error)
}
