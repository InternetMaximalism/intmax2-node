package deposit_synchronizer

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/utils"
	"math/big"
	"sort"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
)

const (
	int32Key  = 32
	int128Key = 128
)

var ErrStatCurrentFileFail = errors.New("stat current file fail")

type depositSynchronizer struct {
	cfg                       *configs.Config
	log                       logger.Logger
	dbApp                     SQLDriverApp
	lastSeenScrollBlockNumber uint64
}

func New(cfg *configs.Config, log logger.Logger, dbApp SQLDriverApp) *depositSynchronizer {
	const startScrollBlockNumber uint64 = 5691248
	return &depositSynchronizer{
		cfg:                       cfg,
		log:                       log,
		dbApp:                     dbApp,
		lastSeenScrollBlockNumber: startScrollBlockNumber,
	}
}

func (w *depositSynchronizer) Init() error {
	return nil
}

func (w *depositSynchronizer) Start(
	ctx context.Context,
	tickerEventWatcher *time.Ticker,
) error {
	rollupCfg := intMaxTypes.NewRollupContractConfigFromEnv(w.cfg, "https://sepolia-rpc.scroll.io")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tickerEventWatcher.C:
			shouldProcess := func() (bool, error) {
				/*
					// latestBlock, err := intMaxTypes.FetchLatestIntMaxBlock(rollupCfg, ctx)
					// if err != nil {
					// 	// "no posted blocks found"のときはエラーを返さない
					// 	if err.Error() != "no posted blocks found" {
					// 		return false, err
					// 	}

					// 	return true, nil
					// }
					// latestDepositTreeRoot, err := intMaxTypes.FetchDepositRoot(rollupCfg, ctx)
					// if err != nil {
					// 	return false, err
					// }

					// if latestBlock.DepositTreeRoot == latestDepositTreeRoot {
					// 	return false, nil
					// }
				*/

				return true, nil
			}

			ok, err := shouldProcess()
			if err != nil {
				return err
			}
			if !ok {
				fmt.Printf("No new deposits\n")
				continue
			}

			// Generate a new block to reflect new deposits.
			// This block includes the transaction of a random generated address.
			// TODO: If there is a block already in the process of being created, there is no need to post this block.
			keyPairs := make([]*intMaxAcc.PrivateKey, 1)
			for i := 0; i < len(keyPairs); i++ {
				var privateKey *big.Int
				privateKey, err = rand.Int(rand.Reader, new(big.Int).Sub(fr.Modulus(), big.NewInt(1)))
				if err != nil {
					return err
				}

				privateKey.Add(privateKey, big.NewInt(1))
				keyPairs[i], err = intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(privateKey)
				if err != nil {
					return err
				}
			}

			// Sort by x-coordinate of public key
			sort.Slice(keyPairs, func(i, j int) bool {
				return keyPairs[i].Pk.X.Cmp(&keyPairs[j].Pk.X) > 0
			})

			senders := make([]intMaxTypes.Sender, int128Key)
			for i, keyPair := range keyPairs {
				senders[i] = intMaxTypes.Sender{
					PublicKey: keyPair.Public(),
					AccountID: 0,
					IsSigned:  true,
				}
			}

			defaultPublicKey := intMaxAcc.NewDummyPublicKey()
			for i := len(keyPairs); i < len(senders); i++ {
				senders[i] = intMaxTypes.Sender{
					PublicKey: defaultPublicKey,
					AccountID: 0,
					IsSigned:  false,
				}
			}

			txRoot, err := new(intMaxTypes.PoseidonHashOut).SetRandom()
			if err != nil {
				return err
			}

			senderPublicKeysBytes := make([]byte, len(senders)*intMaxTypes.NumPublicKeyBytes)
			for i, sender := range senders {
				if sender.IsSigned {
					senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
					copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+1)], senderPublicKey[:])
				}
			}

			publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
			aggregatedPublicKey := new(intMaxAcc.PublicKey)
			for _, sender := range senders {
				if sender.IsSigned {
					aggregatedPublicKey.Add(aggregatedPublicKey, sender.PublicKey.WeightByHash(publicKeysHash))
				}
			}

			message := finite_field.BytesToFieldElementSlice(txRoot.Marshal())

			aggregatedSignature := new(bn254.G2Affine)
			for i, keyPair := range keyPairs {
				if senders[i].IsSigned {
					var signature *bn254.G2Affine
					signature, err = keyPair.WeightByHash(publicKeysHash).Sign(message)
					if err != nil {
						return err
					}
					aggregatedSignature.Add(aggregatedSignature, signature)
				}
			}

			blockContent := intMaxTypes.NewBlockContent(
				intMaxTypes.PublicKeySenderType,
				senders,
				*txRoot,
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

			tx, err := intMaxTypes.PostRegistrationBlock(rollupCfg, blockContent)
			if err != nil {
				return err
			}

			fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
		}
	}
}

func SubscribeDepositsProcessed(cfg *intMaxTypes.RollupContractConfig, ctx context.Context) (eventChan chan *bindings.RollupDepositsProcessed, subscription event.Subscription, err error) {
	client, err := utils.NewClient(cfg.NetworkRpcUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new client: %w", err)
	}
	defer client.Close()

	rollup, err := bindings.NewRollup(common.HexToAddress(cfg.RollupContractAddressHex), client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to instantiate a Liquidity contract: %w", err)
	}

	opts := &bind.WatchOpts{Context: context.Background()}
	eventChan = make(chan *bindings.RollupDepositsProcessed)

	subscription, err = rollup.WatchDepositsProcessed(opts, eventChan, []*big.Int{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to event: %w", err)
	}

	return eventChan, subscription, nil
}
