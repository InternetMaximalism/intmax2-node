package configs

type Wallet struct {
	MnemonicValue          string `env:"WALLET_MNEMONIC_VALUE"`
	MnemonicDerivationPath string `env:"WALLET_MNEMONIC_DERIVATION_PATH"`
	MnemonicPassword       string `env:"WALLET_MNEMONIC_PASSWORD"`

	PrivateKeyHex string `env:"WALLET_PRIVATE_KEY_HEX"`
}
