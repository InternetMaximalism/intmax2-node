package block_post_service

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=block_post_service_test -source=blockchain_service.go

type ServiceBlockchain interface {
	ChainSB
}

type ChainSB interface {
	SetupScrollNetworkChainID(ctx context.Context) error
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}
