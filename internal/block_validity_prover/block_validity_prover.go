package block_validity_prover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var ErrStatCurrentFileFail = errors.New("stat current file fail")

type blockValidityProver struct {
	cfg                       *configs.Config
	log                       logger.Logger
	dbApp                     SQLDriverApp
	lastSeenScrollBlockNumber uint64
	accountInfoMap            block_post_service.AccountInfoMap
}

func New(cfg *configs.Config, log logger.Logger, dbApp SQLDriverApp) BlockValidityProver {
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
			// d, err := block_post_service.NewBlockPostService(ctx, w.cfg)
			// if err != nil {
			// 	return err
			// }

			// events, _, err := d.FetchNewPostedBlocks(w.lastSeenScrollBlockNumber)
			// if err != nil {
			// 	return err
			// }

			// latestBlockNumber, err := d.FetchLatestBlockNumber(ctx)
			// if err != nil {
			// 	return err
			// }

			// if len(events) == 0 {
			// 	w.lastSeenScrollBlockNumber = latestBlockNumber
			// 	continue
			// }

			// lastSeenBlockNumber := w.lastSeenScrollBlockNumber
			// for _, event := range events {
			// 	if event.Raw.BlockNumber > lastSeenBlockNumber {
			// 		lastSeenBlockNumber = event.Raw.BlockNumber
			// 	}

			// 	var calldata []byte
			// 	calldata, err = d.FetchScrollCalldataByHash(event.Raw.TxHash)
			// 	if err != nil {
			// 		continue
			// 	}

			// 	_, err = block_post_service.FetchIntMaxBlockContentByCalldata(calldata, w.accountInfoMap)
			// 	if err != nil {
			// 		if errors.Is(err, block_post_service.ErrUnknownAccountID) {
			// 			continue
			// 		}
			// 		if errors.Is(err, block_post_service.ErrCannotDecodeAddress) {
			// 			continue
			// 		}

			// 		continue
			// 	}
			// }

			// w.lastSeenScrollBlockNumber = lastSeenBlockNumber

			rollupCfg := intMaxTypes.NewRollupContractConfigFromEnv(w.cfg, "https://sepolia-rpc.scroll.io")

			// Post unprocessed block
			unprocessedBlocks, err := w.dbApp.GetUnprocessedBlocks()
			if err != nil {
				return err
			}
			if len(unprocessedBlocks) == 0 {
				fmt.Printf("No unprocessed blocks\n")
				continue
			}

			w.log.Infof("Unprocessed blocks: %d\n", len(unprocessedBlocks))
			for _, unprocessedBlock := range unprocessedBlocks {
				var senderType string
				if unprocessedBlock.SenderType == 0 {
					senderType = "PUBLIC_KEY"
				} else {
					senderType = "ACCOUNT_ID"
				}

				var qSenders []intMaxTypes.ColumnSender
				err := json.Unmarshal(unprocessedBlock.Senders, &qSenders)
				if err != nil {
					return err
				}

				senders := make([]intMaxTypes.Sender, 0)
				for _, sender := range qSenders {
					publicKey, err := intMaxAcc.NewPublicKeyFromAddressHex(sender.PublicKey)
					if err != nil {
						return err
					}

					sender := intMaxTypes.Sender{
						PublicKey: publicKey,
						AccountID: sender.AccountID,
						IsSigned:  sender.IsSigned,
					}
					senders = append(senders, sender)
				}

				txTreeRootBytes, err := hexutil.Decode("0x" + unprocessedBlock.TxRoot)
				if err != nil {
					return err
				}

				txTreeRoot := new(goldenposeidon.PoseidonHashOut)
				err = txTreeRoot.Unmarshal(txTreeRootBytes)
				if err != nil {
					return err
				}

				aggregatedSignatureHex, err := hexutil.Decode("0x" + unprocessedBlock.AggregatedSignature)
				if err != nil {
					return err
				}
				aggregatedSignature := new(bn254.G2Affine)
				err = aggregatedSignature.Unmarshal(aggregatedSignatureHex)
				if err != nil {
					return err
				}

				blockContent := intMaxTypes.NewBlockContent(
					senderType,
					senders,
					*txTreeRoot,
					aggregatedSignature,
				)
				if err = blockContent.IsValid(); err != nil {
					return err
				}

				_, err = intMaxTypes.MakePostRegistrationBlockInput(
					blockContent,
				)
				if err != nil {
					return err
				}

				tx, err := intMaxTypes.PostRegistrationBlock(rollupCfg, blockContent)
				if err != nil {
					return err
				}

				fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())

				err = w.dbApp.UpdateBlockStatus(unprocessedBlock.ProposalBlockID, 1)
				if err != nil {
					return err
				}

				fmt.Println("UpdateBlockStatus")
			}
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
