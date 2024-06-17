package server

import (
	"context"
	"math/big"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_blockchain_service.go -package=server -source=blockchain_service.go

type ServiceBlockchain interface {
	GenericCommandsSB
	WriteBlockchain
	ChainSB
	WalletSB
	RollupSB
}

type GenericCommandsSB interface {
	Check(ctx context.Context) (res health.Health)
	CheckPrivateKey(ctx context.Context) (err error)
}

type ChainSB interface {
	SetupScrollNetworkChainID(ctx context.Context) error
	ScrollNetworkChainLinkEvmJSONRPC(ctx context.Context) (string, error)
}

type WalletSB interface {
	WalletBalance(
		ctx context.Context,
		address common.Address,
	) (bal *big.Int, err error)
}

type RollupSB interface {
	BlockBuilderUrl(ctx context.Context) (string, error)
}
