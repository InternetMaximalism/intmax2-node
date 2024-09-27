package deposit_tree_proof_by_deposit_index

import "context"

//go:generate mockgen -destination=../mocks/mock_deposit_tree_proof_by_deposit_index.go -package=mocks -source=deposit_tree_proof_by_deposit_index.go

type UCDepositTreeProofByDepositIndexInput struct {
	DepositIndex int64 `json:"depositIndex"`
}

type UCDepositTreeProofByDepositIndexMerkleProof struct {
	Siblings []string `json:"siblings"`
}

type UCDepositTreeProofByDepositIndex struct {
	MerkleProof *UCDepositTreeProofByDepositIndexMerkleProof `json:"merkleProof"`
	RootHash    string                                       `json:"rootHash"`
}

type UseCaseDepositTreeProofByDepositIndex interface {
	Do(ctx context.Context, input *UCDepositTreeProofByDepositIndexInput) (*UCDepositTreeProofByDepositIndex, error)
}
