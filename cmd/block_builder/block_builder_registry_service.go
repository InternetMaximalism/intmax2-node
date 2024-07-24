package block_builder

import (
	"context"
	"intmax2-node/internal/block_builder_registry_service"
)

//go:generate mockgen -destination=mock_block_builder_registry_service.go -package=block_builder -source=block_builder_registry_service.go

type BlockBuilderRegistryService interface {
	GetBlockBuilder(
		ctx context.Context,
	) (*block_builder_registry_service.IBlockBuilderRegistryBlockBuilderInfo, error)
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
