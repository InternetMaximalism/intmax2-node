package pow

import "context"

type PoWNonce interface {
	Nonce(ctx context.Context, msg []byte) (nonce string, err error)
	Verify(nonce string, msg []byte) error
}
