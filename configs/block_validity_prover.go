package configs

import "time"

type BlockValidityProver struct {
	TimeoutForFetchingBlockValidityProof time.Duration `env:"BLOCK_VALIDITY_PROVER_FETCH_PROOF_LIFETIME" envDefault:"3s"`
	BlockValidityProverUrl               string        `env:"BLOCK_VALIDITY_PROVER_URL" envDefault:"http://localhost:8091"`
	BalanceValidityProverUrl             string        `env:"BALANCE_VALIDITY_PROVER_URL" envDefault:"http://localhost:8092"`
	WithdrawalProverUrl                  string        `env:"WITHDRAWAL_PROVER_URL" envDefault:"http://localhost:8093"`
}
