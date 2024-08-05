package deposit_service

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=deposit_service_test -source=blockchain_service.go

type ServiceBlockchain interface {
	ChainSB
}

type ChainSB interface {
	SetupEthereumNetworkChainID(ctx context.Context) error
	EthereumNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}
