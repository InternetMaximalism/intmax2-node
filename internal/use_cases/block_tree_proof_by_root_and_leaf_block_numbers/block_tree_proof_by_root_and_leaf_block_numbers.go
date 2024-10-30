package block_tree_proof_by_root_and_leaf_block_numbers

import "context"

//go:generate mockgen -destination=../mocks/mock_block_tree_proof_by_root_and_leaf_block_numbers.go -package=mocks -source=block_tree_proof_by_root_and_leaf_block_numbers.go

type UCBlockTreeProofByRootAndLeafBlockNumbersInput struct {
	RootBlockNumber int64 `json:"rootBlockNumber"`
	LeafBlockNumber int64 `json:"leafBlockNumber"`
}

type UCBlockTreeProofByRootAndLeafBlockNumbersMerkleProof struct {
	Siblings []string `json:"siblings"`
}

type UCBlockTreeProofByRootAndLeafBlockNumbers struct {
	MerkleProof *UCBlockTreeProofByRootAndLeafBlockNumbersMerkleProof `json:"merkleProof"`
	RootHash    string                                                `json:"rootHash"`
}

type UseCaseBlockTreeProofByRootAndLeafBlockNumbers interface {
	Do(ctx context.Context, input *UCBlockTreeProofByRootAndLeafBlockNumbersInput) (*UCBlockTreeProofByRootAndLeafBlockNumbers, error)
}
