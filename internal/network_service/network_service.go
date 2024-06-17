package network_service

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/open_telemetry"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/dimiro1/health"
	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-validation"
)

var NodeExternalAddress struct {
	sync.Mutex
	Address ExternalAddress
}

type networkService struct {
	cfg *configs.Config
}

func New(cfg *configs.Config) NetworkService {
	return &networkService{
		cfg: cfg,
	}
}

func (ns *networkService) networkValidator() (err error) {
	return validation.ValidateStruct(&ns.cfg.Network,
		validation.Field(&ns.cfg.Network.Domain, validation.Required, validation.By(func(value interface{}) error {
			ipOrDomain, ok := value.(string)
			if !ok {
				return ErrValueInvalid
			}

			if net.ParseIP(ipOrDomain) == nil &&
				validation.Validate(ipOrDomain, validation.Required, validation.By(func(value interface{}) error {
					const mockEmailValidator = "mock@"
					val, ok := value.(string)
					if !ok {
						return ErrValueInvalid
					}

					if v := ev.NewSyntaxValidator().Validate(ev.NewInput(evmail.FromString(mockEmailValidator + val))); !v.IsValid() {
						return ErrValueInvalid
					}

					return nil
				})) != nil {
				return ErrValueInvalid
			}

			return nil
		})),
		validation.Field(&ns.cfg.Network.Port, validation.Required, validation.By(func(value interface{}) error {
			port, ok := value.(int)
			if !ok {
				return ErrValueInvalid
			}

			const int0Key = 0
			if port <= int0Key {
				return ErrValueInvalid
			}

			return nil
		})),
	)
}

func (ns *networkService) Check(ctx context.Context) (res health.Health) {
	const (
		externalAddress = "external_address"
		addressType     = "address_type"
		emptyKey        = ""
		maskAddr        = "%s:%d"
	)
	res.AddInfo(externalAddress, emptyKey)
	res.AddInfo(addressType, emptyKey)

	var (
		err error
		ea  ExternalAddress
	)
	ea, err = ns.ExternalAddress(ctx)
	if err != nil {
		res.Down()
		return res
	}

	if NodeExternalAddress.Address.Domain() == emptyKey &&
		NodeExternalAddress.Address.Domain() == NodeExternalAddress.Address.IP().String() {
		res.Down()
		return res
	}

	ned := NodeExternalAddress.Address.Domain()
	if NodeExternalAddress.Address.IP() != nil {
		ned = NodeExternalAddress.Address.IP().String()
	}

	if NodeExternalAddress.Address.Type() != HandInputExternalAddressType &&
		NodeExternalAddress.Address.Type() != EnvInitExternalAddressType &&
		!strings.EqualFold(ned, ea.IP().String()) {
		res.AddInfo(externalAddress, ea.IP().String())
		res.Down()
		return res
	}

	res.AddInfo(externalAddress, fmt.Sprintf(maskAddr, ned, NodeExternalAddress.Address.Port()))
	res.AddInfo(addressType, NodeExternalAddress.Address.Type())
	res.Up()

	return res
}

func (ns *networkService) CheckNetwork(ctx context.Context) (err error) {
	const (
		hName = "NetworkService func:CheckNetwork"

		emptyKey = ""
		int0Key  = 0
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	err = ns.networkValidator()
	if err != nil {
		var extAddr ExternalAddress
		extAddr, err = ns.GatewayAddress(spanCtx)
		if err != nil {
			const (
				enterMSGIpOrDomain = "Enter the external IP or Domain:"
				enterMSGPort       = "Enter the external Port:"
				enterIsHttps       = "Use HTTPS scheme (If yes, enter 'yes'; if no, enter something) ?"
				yesKey             = "yes"
				crlf               = '\n'
			)

			fmt.Printf(enterMSGIpOrDomain)
			var ipOrDomain string
			ipOrDomain, err = bufio.NewReader(os.Stdin).ReadString(crlf)
			if err != nil {
				open_telemetry.MarkSpanError(spanCtx, err)
				return errors.Join(ErrStdinProcessingFail, err)
			}
			ipOrDomain = strings.TrimSpace(ipOrDomain)

			if net.ParseIP(ipOrDomain) == nil &&
				validation.Validate(ipOrDomain, validation.Required, validation.By(func(value interface{}) error {
					const mockEmailValidator = "mock@"
					val, ok := value.(string)
					if !ok {
						return ErrValueInvalid
					}

					if v := ev.NewSyntaxValidator().Validate(ev.NewInput(evmail.FromString(mockEmailValidator + val))); !v.IsValid() {
						return ErrValueInvalid
					}

					return nil
				})) != nil {
				return errors.Join(ErrNetworkIpOrDomainInvalid, err)
			}
			ns.cfg.Network.Domain = ipOrDomain
			ip := net.ParseIP(ipOrDomain)

			fmt.Printf(enterMSGPort)
			var port string
			port, err = bufio.NewReader(os.Stdin).ReadString(crlf)
			if err != nil {
				open_telemetry.MarkSpanError(spanCtx, err)
				return errors.Join(ErrStdinProcessingFail, err)
			}
			port = strings.TrimSpace(port)
			var extPort int
			extPort, err = strconv.Atoi(port)
			if err != nil {
				return errors.Join(ErrStrConvAtoIFail, err)
			}
			ns.cfg.Network.Port = extPort

			fmt.Printf(enterIsHttps)
			var isHttps string
			isHttps, err = bufio.NewReader(os.Stdin).ReadString(crlf)
			if err != nil {
				open_telemetry.MarkSpanError(spanCtx, err)
				return errors.Join(ErrStdinProcessingFail, err)
			}
			if strings.EqualFold(yesKey, strings.TrimSpace(isHttps)) {
				ns.cfg.Network.HTTPSUse = true
			}

			NodeExternalAddress.Lock()
			defer NodeExternalAddress.Unlock()
			if ip != nil {
				NodeExternalAddress.Address = NewExternalAddress(
					ip,
					emptyKey,
					ns.cfg.Network.Port,
					int0Key,
					HandInputExternalAddressType,
					ns.cfg.Network.HTTPSUse,
				)
			} else {
				NodeExternalAddress.Address = NewExternalAddress(
					nil,
					ns.cfg.Network.Domain,
					ns.cfg.Network.Port,
					int0Key,
					HandInputExternalAddressType,
					ns.cfg.Network.HTTPSUse,
				)
			}

			return nil
		}

		ns.cfg.Network.Domain = extAddr.IP().String()
		ns.cfg.Network.Port = extAddr.Port()

		NodeExternalAddress.Lock()
		defer NodeExternalAddress.Unlock()
		NodeExternalAddress.Address = extAddr
		return nil
	}

	NodeExternalAddress.Lock()
	defer NodeExternalAddress.Unlock()
	NodeExternalAddress.Address = NewExternalAddress(
		nil,
		ns.cfg.Network.Domain,
		ns.cfg.Network.Port,
		int0Key,
		EnvInitExternalAddressType,
		ns.cfg.Network.HTTPSUse,
	)

	return nil
}
