package http_request_modifier

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

type CookiesMD map[string]string

func GetCookie(ctx context.Context, name string) (*http.Cookie, error) {
	key := strings.ToLower(Cookie)
	m, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(m[key]) == 0 {
		return nil, nil
	}

	header := http.Header{}
	header.Add(Cookie, m[key][0])
	request := http.Request{Header: header}

	return request.Cookie(name)
}

func GetHeader(ctx context.Context, name string) string {
	key := strings.ToLower(name)
	m, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(m[key]) == 0 {
		const emptyKey = ""
		return emptyKey
	}

	return strings.TrimSpace(m[key][0])
}
