package configs

import "time"

const (
	blockValidityProverMaxValueOfInputDepositsInRequest = 10000
)

type BlockValidityProver struct {
	TimeoutForEventWatcher               time.Duration `env:"BLOCK_VALIDITY_PROVER_EVENT_WATCHER_LIFETIME" envDefault:"1m"`
	TimeoutForFetchingBlockValidityProof time.Duration `env:"BLOCK_VALIDITY_PROVER_FETCH_PROOF_LIFETIME" envDefault:"1m"`
	BlockValidityProverUrl               string        `env:"API_BLOCK_VALIDITY_PROVER_URL" envDefault:"http://localhost:8091"`

	BlockValidityProverMaxValueOfInputDepositsInRequest int `env:"BLOCK_VALIDITY_PROVER_MAX_VALUE_OF_INPUT_DEPOSITS_IN_REQUEST" envDefault:"10000"`
}
