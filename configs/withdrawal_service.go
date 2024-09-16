package configs

type WithdrawalService struct {
	WithdrawalProverUrl      string `env:"WITHDRAWAL_PROVER_URL,required" envDefault:"http://localhost:8093"`
	WithdrawalGnarkProverUrl string `env:"WITHDRAWAL_GNARK_PROVER_URL,required"`
}
