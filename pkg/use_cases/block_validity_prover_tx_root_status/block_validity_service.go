package block_validity_prover_tx_root_status

import (
	"intmax2-node/internal/block_validity_prover"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_block_validity_prover_tx_root_status_test.go -package=block_validity_prover_tx_root_status_test -source=block_validity_service.go

type BlockValidityService interface {
	AuxInfoListFromBlockContentByTxRoot(txRoot ...common.Hash) (map[common.Hash]*block_validity_prover.AuxInfo, error)
}
