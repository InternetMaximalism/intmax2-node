package configs

import "math/big"

type Blockchain struct {
	ScrollNetworkChainID                         string  `env:"BLOCKCHAIN_SCROLL_NETWORK_CHAIN_ID"`
	ScrollNetworkMinBalance                      big.Int `env:"BLOCKCHAIN_SCROLL_MIN_BALANCE" envDefault:"100000000000000000"`
	ScrollNetworkStakeBalance                    big.Int `env:"BLOCKCHAIN_SCROLL_STAKE_BALANCE" envDefault:"100000000000000000"`
	ScrollMessengerL1ContractAddress             string  `env:"BLOCKCHAIN_SCROLL_MESSENGER_L1_CONTRACT_ADDRESS"`
	ScrollMessengerL1ContractDeployedBlockNumber uint64  `env:"BLOCKCHAIN_SCROLL_MESSENGER_L1_CONTRACT_DEPLOYED_BLOCK_NUMBER" envDefault:"0"`
	ScrollMessengerL2ContractAddress             string  `env:"BLOCKCHAIN_SCROLL_MESSENGER_L2_CONTRACT_ADDRESS"`

	RollupContractAddress             string `env:"BLOCKCHAIN_ROLLUP_CONTRACT_ADDRESS,required"`
	RollupContractDeployedBlockNumber uint64 `env:"BLOCKCHAIN_ROLLUP_CONTRACT_DEPLOYED_BLOCK_NUMBER" envDefault:"0"`

	EthereumNetworkChainID string `env:"BLOCKCHAIN_ETHEREUM_NETWORK_CHAIN_ID"`
	EthereumNetworkRpcUrl  string `env:"BLOCKCHAIN_ETHEREUM_NETWORK_RPC_URL"`

	BlockBuilderRegistryContractAddress  string `env:"BLOCKCHAIN_BLOCK_BUILDER_REGISTRY_CONTRACT_ADDRESS"`
	LiquidityContractAddress             string `env:"BLOCKCHAIN_LIQUIDITY_CONTRACT_ADDRESS"`
	LiquidityContractDeployedBlockNumber uint64 `env:"BLOCKCHAIN_LIQUIDITY_CONTRACT_DEPLOYED_BLOCK_NUMBER" envDefault:"0"`
	WithdrawalContractAddress            string `env:"BLOCKCHAIN_WITHDRAWAL_CONTRACT_ADDRESS"`

	MaxCounterOfTransaction int `env:"BLOCKCHAIN_MAX_COUNTER_OF_TRANSACTION" envDefault:"128"`

	BuilderPrivateKeyHex         string `env:"BLOCKCHAIN_ETHEREUM_BUILDER_KEY_HEX"`
	DepositAnalyzerPrivateKeyHex string `env:"BLOCKCHAIN_ETHEREUM_DEPOSIT_ANALYZER_PRIVATE_KEY_HEX"`
	WithdrawalPrivateKeyHex      string `env:"BLOCKCHAIN_ETHEREUM_WITHDRAWAL_PRIVATE_KEY_HEX"`
	MessengerMockPrivateKeyHex   string `env:"BLOCKCHAIN_ETHEREUM_MESSENEGER_MOCK_PRIVATE_KEY_HEX"`

	DepositAnalyzerThreshold        uint64 `env:"BLOCKCHAIN_DEPOSIT_ANALYZER_THRESHOLD" envDefault:"10"`
	DepositAnalyzerMinutesThreshold uint64 `env:"BLOCKCHAIN_DEPOSIT_ANALYZER_MINUTES_THRESHOLD" envDefault:"10"`

	WithdrawalAggregatorThreshold        uint64 `env:"BLOCKCHAIN_WITHDRAWAL_AGGREGATOR_THRESHOLD" envDefault:"8"`
	WithdrawalAggregatorMinutesThreshold uint64 `env:"BLOCKCHAIN_WITHDRAWAL_AGGREGATOR_MINUTES_THRESHOLD" envDefault:"15"`
}
