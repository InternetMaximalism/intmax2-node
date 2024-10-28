package block_validity_prover

import (
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_block_validity_service.go -package=block_validity_prover -source=block_validity_service.go

type BlockValidityService interface {
	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (
		*intMaxTree.PoseidonMerkleProof,
		*intMaxTree.PoseidonHashOut,
		error,
	)
	FetchValidityProverInfo() (*block_validity_prover.ValidityProverInfo, error)
}
