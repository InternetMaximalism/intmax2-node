package balance_prover_service

import (
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/intmax_block_content"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_block_validity_service_test.go -package=balance_prover_service_test -source=block_validity_service.go

type BlockValidityService interface {
	BlockContentByTxRoot(txRoot common.Hash) (*intmax_block_content.PostedBlock, error)
	GetDepositInfoByHash(depositHash common.Hash) (depositInfo *block_validity_prover.DepositInfo, err error)
}
