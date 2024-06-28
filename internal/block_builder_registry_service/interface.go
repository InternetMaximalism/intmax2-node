package block_builder_registry_service

import (
	"context"
	"intmax2-node/internal/bindings"
)

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
