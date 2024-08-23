package block_validity_prover

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"math/big"
	"time"
)

type BlockValidityProver interface {
	Init(ctx context.Context) error
	Start(
		ctx context.Context,
		tickerEventWatcher *time.Ticker,
	) error
	FetchAccountIDFromPublicKey(publicKey *intMaxAcc.PublicKey) (accountID uint64, err error)
	FetchPublicKeyFromAddress(accountID uint64) (publicKey *intMaxAcc.PublicKey, err error)
	FetchDepositMerkleProofFromDepositID(depositID *big.Int) (depositMerkleProof []string, err error)
}
