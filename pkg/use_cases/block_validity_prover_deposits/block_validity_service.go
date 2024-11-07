package block_validity_prover_deposits

import (
	"intmax2-node/internal/block_validity_prover"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_block_validity_prover_deposits_test.go -package=block_validity_prover_deposits_test -source=block_validity_service.go

type BlockValidityService interface {
	GetDepositsInfoByHash(depositHash ...common.Hash) (map[uint32]*block_validity_prover.DepositInfo, error)
}
