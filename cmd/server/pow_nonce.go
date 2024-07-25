package server

import "context"

//go:generate mockgen -destination=mock_pow_nonce.go -package=server -source=pow_nonce.go

type PoWNonce interface {
	Nonce(ctx context.Context, msg []byte) (nonce string, err error)
	Verify(nonce string, msg []byte) error
}
