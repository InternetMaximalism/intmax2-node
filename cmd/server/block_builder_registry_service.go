package server

import "context"

//go:generate mockgen -destination=mock_block_builder_registry_service.go -package=server -source=block_builder_registry_service.go

type BlockBuilderRegistryService interface {
	UpdateBlockBuilder(
		ctx context.Context,
		url string,
	) error
}
