package store_vault_server

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service.go -package=store_vault_server -source=blockchain_service.go

type ServiceBlockchain interface {
	ChainSB
}

type ChainSB interface {
	SetupScrollNetworkChainID(ctx context.Context) error
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}
