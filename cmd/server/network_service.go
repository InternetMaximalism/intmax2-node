package server

import (
	"context"
	"intmax2-node/internal/network_service"

	"github.com/dimiro1/health"
)

//go:generate mockgen -destination=mock_network_service.go -package=server -source=network_service.go

type NetworkService interface {
	GenericCommandsNS
	AddressCommandsNS
}

type GenericCommandsNS interface {
	Check(ctx context.Context) (res health.Health)
	CheckNetwork(ctx context.Context) (err error)
}

type AddressCommandsNS interface {
	LocalIPv4(ctx context.Context) (ip string, err error)
	ExternalAddress(ctx context.Context) (address network_service.ExternalAddress, err error)
	GatewayAddress(ctx context.Context) (address network_service.ExternalAddress, err error)
	NATDiscover(
		ctx context.Context, internalPort int, extAddr network_service.ExternalAddress,
	) (network_service.ExternalAddress, error)
}
