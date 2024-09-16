package configs

type Api struct {
	WithdrawalProverUrl      string `env:"API_WITHDRAWAL_PROVER_URL"`
	WithdrawalGnarkProverUrl string `env:"API_WITHDRAWAL_GNARK_PROVER_URL"`
	ScrollBridgeUrl          string `env:"API_SCROLL_BRIDGE_URL"`
	BlockBuilderUrl          string `env:"API_BLOCK_BUILDER_URL" envDefault:"http://0.0.0.0"`
	DataStoreVaultUrl        string `env:"API_DATA_STORE_VAULT_URL" envDefault:"http://0.0.0.0"`
	WithdrawalServerUrl      string `env:"API_WITHDRAWAL_SERVER_URL" envDefault:"http://0.0.0.0"`
}
