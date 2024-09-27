package block_validity_prover_info

import (
	"intmax2-node/internal/block_validity_prover"
)

//go:generate mockgen -destination=mock_block_validity_service_test.go -package=block_validity_prover_info_test -source=block_validity_service.go

type BlockValidityService interface {
	FetchValidityProverInfo() (*block_validity_prover.ValidityProverInfo, error)
}
