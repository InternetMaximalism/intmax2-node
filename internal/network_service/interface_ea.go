package network_service

import "net"

type ExternalAddressType string

const (
	DirectExternalAddressType      ExternalAddressType = "direct"
	ProxyExternalAddressType       ExternalAddressType = "proxy"
	NatDiscoverExternalAddressType ExternalAddressType = "nat-discover"
	HandInputExternalAddressType   ExternalAddressType = "hand-input"
	EnvInitExternalAddressType     ExternalAddressType = "env-init"
)

type ExternalAddress interface {
	IP() net.IP
	Domain() string
	Port() int
	InternalPort() int
	Address() string
	Type() ExternalAddressType
}
