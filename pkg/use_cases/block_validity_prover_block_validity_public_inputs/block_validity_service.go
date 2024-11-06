package block_validity_prover_block_validity_public_inputs

import (
	"intmax2-node/internal/block_validity_prover"
)

//go:generate mockgen -destination=mock_block_validity_service_test.go -package=block_validity_prover_block_validity_public_inputs_test -source=block_validity_service.go

type BlockValidityService interface {
	ValidityPublicInputsByBlockNumber(
		blockNumber uint32,
	) (
		validityPublicInputs *block_validity_prover.ValidityPublicInputs,
		senderLeaves []block_validity_prover.SenderLeaf,
		err error,
	)
}
