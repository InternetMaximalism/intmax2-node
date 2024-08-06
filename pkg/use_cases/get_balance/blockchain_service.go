package get_balance

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=get_balance_test -source=blockchain_service.go

type ServiceBlockchain interface {
	ChainSB
}

type ChainSB interface {
	EthereumNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}
