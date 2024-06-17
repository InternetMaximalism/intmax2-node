package network_service

import "errors"

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("value must be valid")

// ErrStunServerDialFail error: failed to dial with stun server.
var ErrStunServerDialFail = errors.New("failed to dial with stun server")

// ErrStunServerBuildFail error: failed to build stun server.
var ErrStunServerBuildFail = errors.New("failed to build stun server")

// ErrStunServerProcessingFail error: failed to processing stun server.
var ErrStunServerProcessingFail = errors.New("failed to processing stun server")

// ErrStunServerDoFail error: failed to do with stun server.
var ErrStunServerDoFail = errors.New("failed to do with stun server")

// ErrStunServerGetFromFail error: failed to get-from stun server.
var ErrStunServerGetFromFail = errors.New("failed to get-from stun server")

// ErrContextCancel error: context was cancelled.
var ErrContextCancel = errors.New("context was cancelled")

// ErrNetInterfaceAddrFail error: failed to get net interface address list.
var ErrNetInterfaceAddrFail = errors.New("failed to get net interface address list")

// ErrNetInterfaceAddrNotFound error: there isn't valid address in net interface address list.
var ErrNetInterfaceAddrNotFound = errors.New("there isn't valid address in net interface address list")

// ErrNATDiscoverFail error: failed to discover NAT.
var ErrNATDiscoverFail = errors.New("failed to discover NAT")

// ErrNATAddMappingFail error: failed to add mapping NAT.
var ErrNATAddMappingFail = errors.New("failed to add mapping NAT")

// ErrNATDiscoverPortInvalid error: NAT discover post must be valid.
var ErrNATDiscoverPortInvalid = errors.New("NAT discover post must be valid")

// ErrDetectLocalIPv4Fail error: failed to detect local IPv4.
var ErrDetectLocalIPv4Fail = errors.New("failed to detect local IPv4")

// ErrDetectExternalAddressFail error: failed to detect external address.
var ErrDetectExternalAddressFail = errors.New("failed to detect external address")

// ErrStrConvAtoIFail error: failed to convert string to integer.
var ErrStrConvAtoIFail = errors.New("failed to convert string to integer")

// ErrNetworkIpOrDomainInvalid error: IP or Domain must be valid.
var ErrNetworkIpOrDomainInvalid = errors.New("IP or Domain must be valid")

// ErrGatewayAddressFail error: failed to get gateway address.
var ErrGatewayAddressFail = errors.New("failed to get gateway address")

// ErrStdinProcessingFail error: stdin processing error occurred.
var ErrStdinProcessingFail = errors.New("stdin processing error occurred")
