package block_validity_prover_balance_update_witness

import (
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
)

//go:generate mockgen -destination=mock_block_validity_service_test.go -package=block_validity_prover_balance_update_witness_test -source=block_validity_service.go

type BlockValidityService interface {
	FetchUpdateWitness(
		publicKey *intMaxAcc.PublicKey,
		currentBlockNumber uint32,
		targetBlockNumber uint32,
		isPrevAccountTree bool,
	) (*block_validity_prover.UpdateWitness, error)
}
