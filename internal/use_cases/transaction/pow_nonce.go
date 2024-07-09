package transaction

import "context"

//go:generate mockgen -destination=mock_pow_nonce_test.go -package=transaction_test -source=pow_nonce.go

type PoWNonce interface {
	Nonce(ctx context.Context, msg []byte) (nonce string, err error)
	Verify(nonce string, msg []byte) error
}
