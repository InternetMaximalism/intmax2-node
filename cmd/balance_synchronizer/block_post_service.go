package server

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"math/big"
	"time"
)

//go:generate mockgen -destination=mock_block_post_service.go -package=server -source=block_post_service.go

type BlockPostService interface {
	Init(ctx context.Context) (err error)
	Start(ctx context.Context, tickerEventWatcher *time.Ticker) error
	FetchAccountIDFromPublicKey(publicKey *intMaxAcc.PublicKey) (accountID uint64, err error)
	FetchPublicKeyFromAddress(accountID uint64) (publicKey *intMaxAcc.PublicKey, err error)
	FetchDepositMerkleProofFromDepositID(depositID *big.Int) (depositMerkleProof []string, err error)
}
