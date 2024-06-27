package configs

import "math/big"

type Blockchain struct {
	ScrollNetworkChainID      string  `env:"BLOCKCHAIN_SCROLL_NETWORK_CHAIN_ID"`
	ScrollNetworkMinBalance   big.Int `env:"BLOCKCHAIN_SCROLL_MIN_BALANCE" envDefault:"100000000000000000"`
	ScrollNetworkStakeBalance big.Int `env:"BLOCKCHAIN_SCROLL_STAKE_BALANCE" envDefault:"100000000000000000"`

	RollupContractAddress      string `env:"BLOCKCHAIN_ROLLUP_CONTRACT_ADDRESS,required"`
	TemplateContractRollupPath string `env:"BLOCKCHAIN_TEMPLATE_CONTRACT_ROLLUP_PATH,required" envDefault:"third_party/contracts/Rollup.json"`

	EthereumNetworkChainID string `env:"BLOCKCHAIN_ETHEREUM_NETWORK_CHAIN_ID"`
	EthereumPrivateKeyHex  string `env:"BLOCKCHAIN_ETHEREUM_PRIVATE_KEY_HEX"`

	LiquidityContractAddress string `env:"BLOCKCHAIN_LIQUIDITY_CONTRACT_ADDRESS"`
}
