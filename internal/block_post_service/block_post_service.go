package block_post_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/open_telemetry"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"intmax2-node/pkg/utils"
	"math/big"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

const BlockPostedEventSignatureID = "0xe27163b76905dc373b4ad854ddc9403bbac659c5f1c5191c39e5a7c44574040a"

var ErrStatCurrentFileFail = errors.New("stat current file fail")

type blockPostService struct {
	cfg                       *configs.Config
	log                       logger.Logger
	dbApp                     SQLDriverApp
	sb                        ServiceBlockchain
	lastSeenScrollBlockNumber uint64
	accountInfoMap            AccountInfo
}

func New(cfg *configs.Config, log logger.Logger, dbApp SQLDriverApp, sb ServiceBlockchain) BlockPostService {
	return &blockPostService{
		cfg:                       cfg,
		log:                       log,
		dbApp:                     dbApp,
		sb:                        sb,
		lastSeenScrollBlockNumber: cfg.Blockchain.RollupContractDeployedBlockNumber,
		accountInfoMap:            NewAccountInfo(dbApp),
	}
}
func (w *blockPostService) Init(ctx context.Context) (err error) {
	const (
		hName = "BlockValidityProver func:Init"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	err = w.sb.SetupScrollNetworkChainID(ctx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrSetupScrollNetworkChainIDFail, err)
	}

	return nil
}

func (w *blockPostService) Start(
	ctx context.Context,
	tickerEventWatcher *time.Ticker,
) (err error) {
	const (
		hName = "BlockValidityProver func:Start"
	)

	spanCtx, span := open_telemetry.Tracer().Start(ctx, hName)
	defer span.End()

	err = w.Init(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(ErrInitFail, err)
	}

	var link string
	link, err = w.sb.ScrollNetworkChainLinkEvmJSONRPC(spanCtx)
	if err != nil {
		open_telemetry.MarkSpanError(spanCtx, err)
		return errors.Join(errorsB.ErrScrollNetworkChainLinkEvmJSONRPCFail, err)
	}

	rollupCfg := intMaxTypes.NewRollupContractConfigFromEnv(w.cfg, link)

	scrollClient, err := utils.NewClient(rollupCfg.NetworkRpcUrl)
	if err != nil {
		return fmt.Errorf("failed to create new client: %w", err)
	}
	defer scrollClient.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tickerEventWatcher.C:
			/*
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
			*/

			// Post unprocessed block
			var unprocessedBlocks []*mDBApp.Block
			unprocessedBlocks, err = w.dbApp.GetUnprocessedBlocks()
			if err != nil {
				return err
			}
			if len(unprocessedBlocks) == 0 {
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
				err = json.Unmarshal(unprocessedBlock.Senders, &qSenders)
				if err != nil {
					return err
				}

				senders := make([]intMaxTypes.Sender, 0)
				for _, sender := range qSenders {
					var publicKey *intMaxAcc.PublicKey
					publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(sender.PublicKey)
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

				var txTreeRootBytes []byte
				txTreeRootBytes, err = hexutil.Decode("0x" + unprocessedBlock.TxRoot)
				if err != nil {
					return err
				}

				txTreeRoot := new(goldenposeidon.PoseidonHashOut)
				err = txTreeRoot.Unmarshal(txTreeRootBytes)
				if err != nil {
					return err
				}

				var aggregatedSignatureHex []byte
				aggregatedSignatureHex, err = hexutil.Decode("0x" + unprocessedBlock.AggregatedSignature)
				if err != nil {
					return err
				}
				aggregatedSignature := new(bn254.G2Affine)
				if innerErr := aggregatedSignature.Unmarshal(aggregatedSignatureHex); innerErr != nil {
					return innerErr
				}

				txRootBytes := [32]byte{}
				copy(txRootBytes[:], txTreeRoot.Marshal())

				blockContent := intMaxTypes.NewBlockContent(
					senderType,
					senders,
					txRootBytes,
					aggregatedSignature,
				)
				if innerErr := blockContent.IsValid(); innerErr != nil {
					return innerErr
				}

				_, err = intMaxTypes.MakePostRegistrationBlockInput(
					blockContent,
				)
				if err != nil {
					return err
				}

				var receipt *types.Receipt
				if blockContent.SenderType == intMaxTypes.PublicKeySenderType {
					var txErr error
					receipt, txErr = intMaxTypes.PostRegistrationBlock(rollupCfg, ctx, w.log, scrollClient, blockContent)
					if txErr != nil {
						return txErr
					}
				} else if blockContent.SenderType == intMaxTypes.AccountIDSenderType {
					var txErr error
					receipt, txErr = intMaxTypes.PostNonRegistrationBlock(rollupCfg, ctx, w.log, scrollClient, blockContent)
					if txErr != nil {
						return txErr
					}
				} else {
					panic("invalid sender type")
				}

				var eventLog *types.Log
				ok := false
				for i := 0; i < len(receipt.Logs); i++ {
					if receipt.Logs[i].Topics[0].Hex() == BlockPostedEventSignatureID {
						eventLog = receipt.Logs[i]
						ok = true
						break
					}
				}

				if !ok {
					return errors.New("BlockPosted event not found")
				}

				var rollup *bindings.Rollup
				rollup, err = bindings.NewRollup(common.HexToAddress(rollupCfg.RollupContractAddressHex), scrollClient)
				if err != nil {
					return fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
				}

				var eventData *bindings.RollupBlockPosted
				eventData, err = rollup.ParseBlockPosted(*eventLog)
				if err != nil {
					return err
				}
				blockNumber := uint32(eventData.BlockNumber.Uint64())

				postedBlock := NewPostedBlock(
					eventData.PrevBlockHash,
					eventData.DepositTreeRoot,
					blockNumber,
					eventData.SignatureHash,
				)

				blockHash := postedBlock.Hash()
				w.log.Infof("INTMAX Block hash: %s\n", blockHash.Hex())

				err = w.dbApp.UpdateBlockStatus(unprocessedBlock.ProposalBlockID, blockHash.Hex(), blockNumber)
				if err != nil {
					return err
				}

				w.log.Infof("Posted registration block. The block number is %d.\n", blockNumber)
			}
		}
	}
}

func (w *blockPostService) FetchAccountIDFromPublicKey(publicKey *intMaxAcc.PublicKey) (accountID uint64, err error) {
	return 0, nil
}

func (w *blockPostService) FetchPublicKeyFromAddress(accountID uint64) (publicKey *intMaxAcc.PublicKey, err error) {
	return nil, nil
}

func (w *blockPostService) FetchDepositMerkleProofFromDepositID(depositID *big.Int) (depositMerkleProof []string, err error) {
	return nil, nil
}
