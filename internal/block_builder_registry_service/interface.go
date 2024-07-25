package block_builder_registry_service

import (
	"context"
	"math/big"
)

type IBlockBuilderRegistryBlockBuilderInfo struct {
	BlockBuilderUrl string
	StakeAmount     *big.Int
	StopTime        *big.Int
	NumSlashes      *big.Int
	IsValid         bool
}

type BlockBuilderRegistryService interface {
	GetBlockBuilder(
		ctx context.Context,
	) (*IBlockBuilderRegistryBlockBuilderInfo, error)
	UpdateBlockBuilder(
		ctx context.Context,
		url string,
	) error
	StopBlockBuilder(
		ctx context.Context,
	) (err error)
	UnStakeBlockBuilder(
		ctx context.Context,
	) (err error)
}
