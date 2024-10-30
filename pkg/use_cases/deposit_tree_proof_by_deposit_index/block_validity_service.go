package deposit_tree_proof_by_deposit_index

import (
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_block_validity_service_test.go -package=deposit_tree_proof_by_deposit_index_test -source=block_validity_service.go

type BlockValidityService interface {
	LatestDepositTreeProofByBlockNumber(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	// DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
}
