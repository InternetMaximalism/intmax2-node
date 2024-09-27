package block_tree_proof_by_root_and_leaf_block_numbers

import (
	intMaxTree "intmax2-node/internal/tree"
)

//go:generate mockgen -destination=mock_block_validity_service_test.go -package=block_tree_proof_by_root_and_leaf_block_numbers_test -source=block_validity_service.go

type BlockValidityService interface {
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (
		*intMaxTree.PoseidonMerkleProof,
		*intMaxTree.PoseidonHashOut,
		error,
	)
}
