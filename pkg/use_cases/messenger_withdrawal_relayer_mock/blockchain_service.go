package messenger_withdrawal_relayer_mock

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=messenger_withdrawal_relayer_mock_test -source=blockchain_service.go

type ServiceBlockchain interface {
	ChainSB
}

type ChainSB interface {
	EthereumNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}
