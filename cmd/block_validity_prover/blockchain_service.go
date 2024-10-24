package block_validity_prover

import (
	"context"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_blockchain_service.go -package=block_validity_prover -source=blockchain_service.go

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
	SetupEthereumNetworkChainID(ctx context.Context) error
	EthereumNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	SetupScrollNetworkChainID(ctx context.Context) error
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	ScrollNetworkChainLinkRollupExplorer(ctx context.Context) (string, error)
}
