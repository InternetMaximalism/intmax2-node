package block_validity_prover_block_status_by_block_number

import "context"

//go:generate mockgen -destination=../mocks/mock_block_validity_prover_block_status_by_block_number.go -package=mocks -source=block_validity_prover_block_status_by_block_number.go

const (
	NotFoundMessage = "Block number not found."
)

type UCBlockValidityProverBlockStatusByBlockNumber struct {
	BlockNumber                 int64  `json:"blockNumber"`
	BlockHash                   string `json:"blockHash"`
	ExecutedBlockHashOnScroll   string `json:"executedBlockHashOnScroll"`
	ExecutedBlockHashOnEthereum string `json:"executedBlockHashOnEthereum"`
}

type UCBlockValidityProverBlockStatusByBlockNumberInput struct {
	BlockNumber uint64 `json:"blockNumber"`
}

type UseCaseBlockValidityProverBlockStatusByBlockNumber interface {
	Do(
		ctx context.Context,
		input *UCBlockValidityProverBlockStatusByBlockNumberInput,
	) (*UCBlockValidityProverBlockStatusByBlockNumber, error)
}
