package http_response_modifier

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/protobuf/proto"
)

type Cookies struct {
	ForAuthUse         bool
	Secure             bool
	Domain             string
	SameSiteStrictMode bool
}

type Processing interface {
	Middleware(
		ctx context.Context,
		w http.ResponseWriter,
		msg proto.Message,
	) (err error)
}

type processing struct {
	cookies *Cookies
}

func NewProcessing(cookies *Cookies) Processing {
	return &processing{
		cookies: cookies,
	}
}

func (p *processing) Middleware(
	ctx context.Context,
	w http.ResponseWriter,
	msg proto.Message,
) (err error) {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	return New(ctx, w, &md, msg, p.cookies).Apply()
}

func Middleware(
	ctx context.Context,
	w http.ResponseWriter,
	msg proto.Message,
	cookies *Cookies,
) (err error) {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}

	return New(ctx, w, &md, msg, cookies).Apply()
}
