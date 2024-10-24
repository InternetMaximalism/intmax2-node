package blockchain

import (
	"context"

	"github.com/dimiro1/health"
)

type ServiceBlockchain interface {
	GenericCommandsSB
	ChainSB
}

type GenericCommandsSB interface {
	Check(ctx context.Context) (res health.Health)
	CheckScrollPrivateKey(ctx context.Context) (err error)
	CheckEthereumPrivateKey(ctx context.Context) (err error)
}

type ChainSB interface {
	SetupScrollNetworkChainID(ctx context.Context) error
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	ScrollNetworkChainLinkExplorer(ctx context.Context) (string, error)
	ScrollNetworkChainLinkRollupExplorer(ctx context.Context) (string, error)
	SetupEthereumNetworkChainID(ctx context.Context) error
	EthereumNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	EthereumNetworkChainLinkExplorer(ctx context.Context) (string, error)
}
