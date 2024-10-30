package block_validity_prover_block_status_by_block_hash

import "context"

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_block_status_by_block_hash.go -package=mocks -source=block_validity_prover_block_status_by_block_hash.go

const (
	NotFoundMessage = "Block hash not found."
)

type UCBlockValidityProverBlockStatusByBlockHash struct {
	BlockNumber                 int64  `json:"blockNumber"`
	BlockHash                   string `json:"blockHash"`
	ExecutedBlockHashOnScroll   string `json:"executedBlockHashOnScroll"`
	ExecutedBlockHashOnEthereum string `json:"executedBlockHashOnEthereum"`
}

type UCBlockValidityProverBlockStatusByBlockHashInput struct {
	BlockHash string `json:"blockHash"`
}

type UseCaseBlockValidityProverBlockStatusByBlockHash interface {
	Do(
		ctx context.Context,
		input *UCBlockValidityProverBlockStatusByBlockHashInput,
	) (*UCBlockValidityProverBlockStatusByBlockHash, error)
}
