package configs

import "math/big"

type Blockchain struct {
	ScrollNetworkChainID             string  `env:"BLOCKCHAIN_SCROLL_NETWORK_CHAIN_ID"`
	ScrollNetworkMinBalance          big.Int `env:"BLOCKCHAIN_SCROLL_MIN_BALANCE" envDefault:"100000000000000000"`
	ScrollNetworkStakeBalance        big.Int `env:"BLOCKCHAIN_SCROLL_STAKE_BALANCE" envDefault:"100000000000000000"`
	ScrollBridgeApiUrl               string  `env:"BLOCKCHAIN_SCROLL_BRIDGE_API_URL"`
	ScrollMessengerL1ContractAddress string  `env:"BLOCKCHAIN_SCROLL_MESSENGER_L1_CONTRACT_ADDRESS"`
	ScrollMessengerL2ContractAddress string  `env:"BLOCKCHAIN_SCROLL_MESSENGER_L2_CONTRACT_ADDRESS"`

	RollupContractAddress      string `env:"BLOCKCHAIN_ROLLUP_CONTRACT_ADDRESS,required"`
	TemplateContractRollupPath string `env:"BLOCKCHAIN_TEMPLATE_CONTRACT_ROLLUP_PATH,required" envDefault:"third_party/contracts/Rollup.json"`

	EthereumNetworkChainID string `env:"BLOCKCHAIN_ETHEREUM_NETWORK_CHAIN_ID"`
	EthereumPrivateKeyHex  string `env:"BLOCKCHAIN_ETHEREUM_PRIVATE_KEY_HEX"`
	EthereumNetworkRpcUrl  string `env:"BLOCKCHAIN_ETHEREUM_NETWORK_RPC_URL"`

	BlockBuilderRegistryContractAddress string `env:"BLOCKCHAIN_BLOCK_BUILDER_REGISTRY_CONTRACT_ADDRESS"`
	LiquidityContractAddress            string `env:"BLOCKCHAIN_LIQUIDITY_CONTRACT_ADDRESS"`
	WithdrawalContractAddress           string `env:"BLOCKCHAIN_WITHDRAWAL_CONTRACT_ADDRESS"`

	DepositAnalyzerPrivateKeyHex string `env:"BLOCKCHAIN_ETHEREUM_DEPOSIT_ANALYZER_PRIVATE_KEY_HEX"`
	DepositRelayerPrivateKeyHex  string `env:"BLOCKCHAIN_ETHEREUM_DEPOSIT_RELAYER_PRIVATE_KEY_HEX"`
	WithdrawalPrivateKeyHex      string `env:"BLOCKCHAIN_ETHEREUM_WITHDRAWAL_PRIVATE_KEY_HEX"`
	MockMessagingPrivateKeyHex   string `env:"BLOCKCHAIN_ETHEREUM_MOCK_MESSAGING_PRIVATE_KEY_HEX"`

	MaxCounterOfTransaction int `env:"BLOCKCHAIN_MAX_COUNTER_OF_TRANSACTION" envDefault:"128"`
}
