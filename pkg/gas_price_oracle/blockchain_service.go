package gas_price_oracle

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=gas_price_oracle_test -source=blockchain_service.go

type ServiceBlockchain interface {
	ChainSB
}

type ChainSB interface {
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}
