package server

import (
	"context"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_blockchain_service.go -package=server -source=blockchain_service.go

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
