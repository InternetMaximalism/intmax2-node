package network_service

import (
	"fmt"
	"net"
	"strings"
)

type externalAddress struct {
	ip           net.IP
	domain       string
	port         int
	internalPort int
	typeOf       ExternalAddressType
	httpsUse     bool
}

func NewExternalAddress(
	ip net.IP,
	domain string,
	port, internalPort int,
	typeOf ExternalAddressType,
	httpsUse bool,
) ExternalAddress {
	return &externalAddress{
		ip:           ip,
		domain:       domain,
		port:         port,
		internalPort: internalPort,
		typeOf:       typeOf,
		httpsUse:     httpsUse,
	}
}

func (e *externalAddress) IP() net.IP {
	return e.ip
}

func (e *externalAddress) Domain() string {
	return e.domain
}

func (e *externalAddress) Port() int {
	return e.port
}

func (e *externalAddress) InternalPort() int {
	return e.internalPort
}

func (e *externalAddress) Address() string {
	const (
		emptyKey   = ""
		maskFormat = "%s%s:%d"
		httpKey    = "http://"
		httpsKey   = "https://"
	)
	schema := httpKey
	if e.httpsUse {
		schema = httpsKey
	}
	if strings.TrimSpace(e.domain) != emptyKey {
		return fmt.Sprintf(maskFormat, schema, e.domain, e.port)
	}

	return fmt.Sprintf(maskFormat, schema, e.ip.String(), e.port)
}

func (e *externalAddress) Type() ExternalAddressType {
	return e.typeOf
}
