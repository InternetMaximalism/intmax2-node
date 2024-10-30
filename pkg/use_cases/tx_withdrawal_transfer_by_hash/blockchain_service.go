package tx_withdrawal_transfer_by_hash

import (
	"context"
)

//go:generate mockgen -destination=mock_blockchain_service_test.go -package=tx_withdrawal_transfer_by_hash_test -source=blockchain_service.go

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
