package block_builder

import (
	"context"
	"intmax2-node/internal/bindings"
)

//go:generate mockgen -destination=mock_block_builder_registry_service.go -package=block_builder -source=block_builder_registry_service.go

type BlockBuilderRegistryService interface {
	GetBlockBuilder(
		ctx context.Context,
	) (*bindings.IBlockBuilderRegistryBlockBuilderInfo, error)
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
