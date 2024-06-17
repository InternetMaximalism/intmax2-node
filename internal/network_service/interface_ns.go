package network_service

import (
	"context"

	"github.com/dimiro1/health"
)

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
	ExternalAddress(ctx context.Context) (address ExternalAddress, err error)
	GatewayAddress(ctx context.Context) (address ExternalAddress, err error)
	NATDiscover(ctx context.Context, internalPort int, extAddr ExternalAddress) (ExternalAddress, error)
}
