package network_service

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/internal/open_telemetry"
	"net"
	"strconv"
	"strings"

	gethNAT "github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/pion/stun"
)

func (ns *networkService) LocalIPv4(ctx context.Context) (string, error) {
	const (
		hName = "NetworkService func:LocalIPv4"

		emptyKey = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	var (
		err  error
		list []net.Addr
	)
	list, err = net.InterfaceAddrs()
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return emptyKey, errors.Join(ErrNetInterfaceAddrFail, err)
	}

	for keyL := range list {
		if ipNet, ok := list[keyL].(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	open_telemetry.MarkSpanError(spanCtx, ErrNetInterfaceAddrNotFound)
	return emptyKey, ErrNetInterfaceAddrNotFound
}

func (ns *networkService) ExternalAddress(ctx context.Context) (ExternalAddress, error) {
	const (
		hName    = "NetworkService func:ExternalAddress"
		emptyKey = ""
		int0Key  = 0
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	var (
		err error
		c   *stun.Client
	)
	for keyL := 0; keyL < len(ns.cfg.StunServer.List); keyL++ {
		for keyN := 0; keyN < len(ns.cfg.StunServer.NetworkType); keyN++ {
			c, err = stun.Dial(ns.cfg.StunServer.NetworkType[keyN], ns.cfg.StunServer.List[keyL])
			if err != nil {
				continue
			}
			break
		}
		if c != nil {
			break
		}
	}
	if c == nil {
		open_telemetry.MarkSpanError(spanCtx, ErrStunServerDialFail)
		return nil, ErrStunServerDialFail
	}

	var message *stun.Message
	message, err = stun.Build(stun.TransactionID, stun.BindingRequest)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrStunServerBuildFail, err)
	}

	clientOut := make(chan stun.Event)
	clientErr := make(chan error)

	go func() {
		err = c.Do(message, func(res stun.Event) {
			clientOut <- res
		})
		if err != nil {
			clientErr <- err
		}
	}()

	select {
	case err = <-clientErr:
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrStunServerProcessingFail, err)
	case res := <-clientOut:
		if res.Error != nil {
			open_telemetry.MarkSpanError(spanCtx, err)
			return nil, errors.Join(ErrStunServerDoFail, err)
		}
		var xorAddr stun.XORMappedAddress
		err = xorAddr.GetFrom(res.Message)
		switch {
		case err != nil:
			var mappedAddr stun.MappedAddress
			err = mappedAddr.GetFrom(res.Message)
			if err != nil {
				open_telemetry.MarkSpanError(spanCtx, err)
				return nil, errors.Join(ErrStunServerGetFromFail, err)
			}

			return NewExternalAddress(
				mappedAddr.IP, emptyKey, mappedAddr.Port, int0Key, ProxyExternalAddressType,
				ns.cfg.Network.HTTPSUse,
			), nil
		default:
			return NewExternalAddress(
				xorAddr.IP, emptyKey, xorAddr.Port, int0Key, ProxyExternalAddressType,
				ns.cfg.Network.HTTPSUse,
			), nil
		}
	case <-spanCtx.Done():
		open_telemetry.MarkSpanError(spanCtx, ErrContextCancel)
		return nil, ErrContextCancel
	}
}

func (ns *networkService) GatewayAddress(ctx context.Context) (address ExternalAddress, err error) {
	const (
		hName    = "NetworkService func:GatewayAddress"
		emptyKey = ""
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	var localIPv4 string
	localIPv4, err = ns.LocalIPv4(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrDetectLocalIPv4Fail, err)
	}

	var extAddr ExternalAddress
	extAddr, err = ns.ExternalAddress(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrDetectExternalAddressFail, err)
	}

	var httpPort int
	httpPort, err = strconv.Atoi(ns.cfg.HTTP.Port)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrStrConvAtoIFail, err)
	}

	if strings.EqualFold(localIPv4, extAddr.IP().String()) {
		return NewExternalAddress(
			extAddr.IP(), emptyKey, httpPort, httpPort, DirectExternalAddressType,
			ns.cfg.Network.HTTPSUse,
		), nil
	}

	var natDExtAddr ExternalAddress
	natDExtAddr, err = ns.NATDiscover(
		spanCtx,
		httpPort,
		extAddr,
	)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrNATDiscoverFail, err)
	}

	return natDExtAddr, nil
}

func (ns *networkService) NATDiscover(
	ctx context.Context,
	internalPort int,
	extAddr ExternalAddress,
) (ExternalAddress, error) {
	const (
		hName  = "NetworkService func:NATDiscover"
		tcpKey = "TCP"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	var (
		err     error
		mapPort uint16
		gn      = gethNAT.Any()
	)
	mapPort, err = gn.AddMapping(
		tcpKey, extAddr.Port(), internalPort, buildvars.BuildName, configs.NatDiscoverLifeTime,
	)

	if err == nil && int(mapPort) == extAddr.Port() {
		return NewExternalAddress(
			extAddr.IP(), extAddr.Domain(), extAddr.Port(), internalPort, NatDiscoverExternalAddressType,
			ns.cfg.Network.HTTPSUse,
		), nil
	}
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return nil, errors.Join(ErrNATAddMappingFail, err)
	}

	open_telemetry.MarkSpanError(spanCtx, ErrNATDiscoverPortInvalid)
	return nil, ErrNATDiscoverPortInvalid
}
