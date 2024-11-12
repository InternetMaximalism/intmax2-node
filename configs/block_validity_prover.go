package configs

import "time"

const (
	blockValidityProverMaxValueOfInputDepositsInRequest = 10000
	blockValidityProverMaxValueOfInputTxRootInRequest   = 10
)

type BlockValidityProver struct {
	TimeoutForEventWatcher               time.Duration `env:"BLOCK_VALIDITY_PROVER_EVENT_WATCHER_LIFETIME" envDefault:"1m"`
	TimeoutForFetchingBlockValidityProof time.Duration `env:"BLOCK_VALIDITY_PROVER_FETCH_PROOF_LIFETIME" envDefault:"1m"`
	BlockValidityProverUrl               string        `env:"API_BLOCK_VALIDITY_PROVER_URL" envDefault:"http://localhost:8091"`

	BlockValidityProverMaxValueOfInputDepositsInRequest int `env:"BLOCK_VALIDITY_PROVER_MAX_VALUE_OF_INPUT_DEPOSITS_IN_REQUEST" envDefault:"10000"`

	BlockValidityProverMaxValueOfInputTxRootInRequest int      `env:"BLOCK_VALIDITY_PROVER_MAX_VALUE_OF_INPUT_TX_ROOT_IN_REQUEST" envDefault:"10"`
	BlockValidityProverInvalidTxRootInRequest         []string `env:"BLOCK_VALIDITY_PROVER_INVALID_TX_ROOT_IN_REQUEST" envSeparator:";" envDefault:"0xfe6fd7720cfd29168d72cff3db0a7a5ad31bd45195f9a9272bd367124a2989b3"`
}
