package withdrawal

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service.go -package=withdrawal -source=blockchain_service.go

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
}
