package blockchain

import (
	"context"
	"math/big"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
)

type ServiceBlockchain interface {
	GenericCommandsSB
	WriteBlockchain
	ChainSB
	WalletSB
}

type GenericCommandsSB interface {
	Check(ctx context.Context) (res health.Health)
	CheckScrollPrivateKey(ctx context.Context) (err error)
	CheckEthereumPrivateKey(ctx context.Context) (err error)
}

type ChainSB interface {
	SetupScrollNetworkChainID(ctx context.Context) error
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	ScrollNetworkChainLinkExplorer(ctx context.Context) (string, error)
	SetupEthereumNetworkChainID(ctx context.Context) error
	EthereumNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
	EthereumNetworkChainLinkExplorer(ctx context.Context) (string, error)
}

type WalletSB interface {
	WalletBalance(
		ctx context.Context,
		address common.Address,
	) (bal *big.Int, err error)
}
