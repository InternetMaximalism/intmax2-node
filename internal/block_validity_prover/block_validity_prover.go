package block_validity_prover

import (
	"context"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/logger"
	"math/big"
	"time"
)

var ErrStatCurrentFileFail = errors.New("stat current file fail")

type blockValidityProver struct {
	cfg                       *configs.Config
	log                       logger.Logger
	dbApp                     SQLDriverApp
	lastSeenScrollBlockNumber uint64
	accountInfoMap            block_post_service.AccountInfoMap
}

func New(cfg *configs.Config, log logger.Logger, dbApp SQLDriverApp) *blockValidityProver {
	return &blockValidityProver{
		cfg:                       cfg,
		log:                       log,
		dbApp:                     dbApp,
		lastSeenScrollBlockNumber: cfg.Blockchain.RollupContractDeployedBlockNumber,
		accountInfoMap:            block_post_service.NewAccountInfoMap(),
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
			d, err := block_post_service.NewBlockPostService(ctx, w.cfg)
			if err != nil {
				return err
			}

			events, _, err := d.FetchNewPostedBlocks(w.lastSeenScrollBlockNumber)
			if err != nil {
				return err
			}

			latestBlockNumber, err := d.FetchLatestBlockNumber(ctx)
			if err != nil {
				return err
			}

			if len(events) == 0 {
				w.lastSeenScrollBlockNumber = latestBlockNumber
				continue
			}

			lastSeenBlockNumber := w.lastSeenScrollBlockNumber
			for _, event := range events {
				if event.Raw.BlockNumber > lastSeenBlockNumber {
					lastSeenBlockNumber = event.Raw.BlockNumber
				}

				calldata, err := d.FetchScrollCalldataByHash(event.Raw.TxHash)
				if err != nil {
					continue
				}

				_, err = block_post_service.FetchIntMaxBlockContentByCalldata(calldata, w.accountInfoMap)
				if err != nil {
					if errors.Is(err, block_post_service.ErrUnknownAccountID) {
						continue
					}
					if errors.Is(err, block_post_service.ErrCannotDecodeAddress) {
						continue
					}

					continue
				}
			}

			w.lastSeenScrollBlockNumber = lastSeenBlockNumber
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
