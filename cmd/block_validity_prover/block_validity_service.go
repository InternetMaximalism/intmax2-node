package block_validity_prover

import (
	"github.com/ethereum/go-ethereum/common"
	intMaxTree "intmax2-node/internal/tree"
)

//go:generate mockgen -destination=mock_block_validity_service.go -package=block_validity_prover -source=block_validity_service.go

type BlockValidityService interface {
	DepositTreeProof(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
}
