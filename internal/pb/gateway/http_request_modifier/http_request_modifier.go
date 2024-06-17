package http_request_modifier

import "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

func Middleware(key string) (string, bool) {
	switch key {
	case Cookie, XChainID, XLocale: // "Authorization" - passing by default
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
