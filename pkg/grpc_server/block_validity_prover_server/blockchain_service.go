package block_validity_prover_server

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=block_validity_prover_server_test -source=blockchain_service.go

type ServiceBlockchain interface {
	ChainSB
}

type ChainSB interface {
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}