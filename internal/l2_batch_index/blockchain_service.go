package l2_batch_index

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=l2_batch_index_test -source=blockchain_service.go

type ServiceBlockchain interface {
	GenericCommandsSB
	ChainSB
}

type GenericCommandsSB interface {
	CheckEthereumPrivateKey(ctx context.Context) (err error)
}

type ChainSB interface {
	SetupEthereumNetworkChainID(ctx context.Context) error
	EthereumNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	SetupScrollNetworkChainID(ctx context.Context) error
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	ScrollNetworkChainLinkRollupExplorer(ctx context.Context) (string, error)
}