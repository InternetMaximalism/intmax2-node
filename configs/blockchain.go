package configs

import "math/big"

type Blockchain struct {
	ScrollNetworkChainID      string  `env:"BLOCKCHAIN_SCROLL_NETWORK_CHAIN_ID"`
	ScrollNetworkMinBalance   big.Int `env:"BLOCKCHAIN_SCROLL_MIN_BALANCE" envDefault:"100000000000000000"`
	ScrollNetworkStakeBalance big.Int `env:"BLOCKCHAIN_SCROLL_STAKE_BALANCE" envDefault:"100000000000000000"`

	RollupContractAddress      string `env:"BLOCKCHAIN_ROLLUP_CONTRACT_ADDRESS,required"`
	TemplateContractRollupPath string `env:"BLOCKCHAIN_TEMPLATE_CONTRACT_ROLLUP_PATH,required" envDefault:"third_party/contracts/Rollup.json"`

	// NOTE: refine following fields
	EthreumNetworkChainID    string `env:"BLOCKCHAIN_ETHREUM_NETWORK_CHAIN_ID"`
	EthreumNetworkRpcURL     string `env:"BLOCKCHAIN_ETHREUM_NETWORK_RPC_URL"`
	LiquidityContractAddress string `env:"BLOCKCHAIN_LIQUIDITY_CONTRACT_ADDRESS"`
	PRIVATE_KEY              string `env:"BLOCKCHAIN_PRIVATE_KEY"`
}
