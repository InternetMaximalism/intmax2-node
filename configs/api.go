package configs

type Api struct {
	ScrollBridgeUrl     string `env:"API_SCROLL_BRIDGE_URL"`
	BlockBuilderUrl     string `env:"API_BLOCK_BUILDER_URL,required" envDefault:"http://0.0.0.0"`
	DataStoreVaultUrl   string `env:"API_DATA_STORE_VAULT_URL,required" envDefault:"http://0.0.0.0"`
	WithdrawalServerUrl string `env:"API_WITHDRAWAL_SERVER_URL,required" envDefault:"http://0.0.0.0"`
}
