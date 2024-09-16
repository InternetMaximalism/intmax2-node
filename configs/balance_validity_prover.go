package configs

type BalanceValidityProver struct {
	BalanceValidityProverUrl string `env:"BALANCE_VALIDITY_PROVER_URL" envDefault:"http://localhost:8092"`
}
