package configs

import "math/big"

type Blockchain struct {
	ScrollNetworkChainID    string  `env:"BLOCKCHAIN_SCROLL_NETWORK_CHAIN_ID"`
	ScrollNetworkMinBalance big.Int `env:"BLOCKCHAIN_SCROLL_MIN_BALANCE" envDefault:"100000000000000000"`

	RollupContractAddress          string `env:"BLOCKCHAIN_ROLLUP_CONTRACT_ADDRESS,required"`
	TemplateContractRollupPath     string `env:"BLOCKCHAIN_TEMPLATE_CONTRACT_ROLLUP_PATH,required" envDefault:"templates/contracts/Rollup.json"`
	EventEncodeBlockBuilderUpdated string `env:"BLOCKCHAIN_EVENT_ENCODE_BLOCK_BUILDER_UPDATED,required" envDefault:"0x"`
	EventNameBlockBuilderUpdated   string `env:"BLOCKCHAIN_EVENT_NAME_BLOCK_BUILDER_UPDATED,required" envDefault:"BlockBuilderUpdated"`
}
