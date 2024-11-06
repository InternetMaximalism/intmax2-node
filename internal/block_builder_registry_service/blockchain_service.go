package block_builder_registry_service

import (
	"context"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=block_builder_registry_service_test -source=blockchain_service.go

type ServiceBlockchain interface {
	GenericCommandsSB
	ChainSB
}

type GenericCommandsSB interface {
	Check(ctx context.Context) (res health.Health)
	CheckScrollPrivateKey(ctx context.Context) (err error)
}

type ChainSB interface {
	SetupScrollNetworkChainID(ctx context.Context) error
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}
