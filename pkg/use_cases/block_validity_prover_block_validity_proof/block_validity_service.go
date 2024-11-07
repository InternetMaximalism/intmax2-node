package block_validity_prover_block_validity_proof

import (
	"intmax2-node/internal/block_validity_prover"
)

//go:generate mockgen -destination=mock_block_validity_service_test.go -package=block_validity_prover_block_validity_proof_test -source=block_validity_service.go

type BlockValidityService interface {
	ValidityProofByBlockNumber(blockNumber uint32) (*block_validity_prover.ValidityProof, error)
}
