package block_validity_prover

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/logger"
	"math/big"
	"time"
)

var ErrStatCurrentFileFail = errors.New("stat current file fail")

type blockValidityProver struct {
	cfg   *configs.Config
	log   logger.Logger
	dbApp SQLDriverApp
}

func New(cfg *configs.Config, log logger.Logger, dbApp SQLDriverApp) *blockValidityProver {
	return &blockValidityProver{
		cfg:   cfg,
		log:   log,
		dbApp: dbApp,
	}
}

func (w *blockValidityProver) Init() error {
	return nil
}

func (w *blockValidityProver) Start(
	ctx context.Context,
	tickerEventWatcher *time.Ticker,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tickerEventWatcher.C:
			// do something
		}
	}
}

func (w *blockValidityProver) FetchAccountIDFromPublicKey(publicKey *intMaxAcc.PublicKey) (accountID uint64, err error) {
	return 0, nil
}

func (w *blockValidityProver) FetchPublicKeyFromAddress(accountID uint64) (publicKey *intMaxAcc.PublicKey, err error) {
	return nil, nil
}

func (w *blockValidityProver) FetchDepositMerkleProofFromDepositID(depositID *big.Int) (depositMerkleProof []string, err error) {
	return nil, nil
}
