package worker_test

import (
	"context"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/mnemonic_wallet/models"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/worker"
	"intmax2-node/pkg/logger"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/iden3/go-iden3-crypto/keccak256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorkerReceiver(t *testing.T) {
	const int2Key = 2
	assert.NoError(t, configs.LoadDotEnv(int2Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()

	var err error

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)

	dbApp := NewMockSQLDriverApp(ctrl)

	w := worker.New(cfg, log, dbApp)

	cfg.Worker.Path = "./mocks/worker"
	cfg.Worker.PathCleanInStart = true

	err = w.Init()
	assert.NoError(t, err)

	tickerCurrentFile := time.NewTicker(cfg.Worker.TimeoutForCheckCurrentFile)
	defer func() {
		if tickerCurrentFile != nil {
			tickerCurrentFile.Stop()
		}
	}()

	tickerSignaturesAvailableFiles := time.NewTicker(cfg.Worker.TimeoutForSignaturesAvailableFiles)
	defer func() {
		if tickerSignaturesAvailableFiles != nil {
			tickerSignaturesAvailableFiles.Stop()
		}
	}()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = w.Start(ctx, tickerCurrentFile, tickerSignaturesAvailableFiles)
		assert.NoError(t, err)
	}()

	const (
		derivation  = "m/44'/60'/0'/0/0"
		userCounter = 1
		emptyKey    = ""
	)
	amount := new(big.Int).SetInt64(int64(1))
	sendersList := make([]*models.Wallet, userCounter)
	recipientsList := make([]*models.Wallet, userCounter)
	assert.NoError(t, err)
	var receiversListForWorker []*worker.ReceiverWorker
	userCounterCheck := userCounter
	dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ interface{}, executor func(d interface{}, input interface{}) error) (err error) {
		userCounterCheck--
		return nil
	}).AnyTimes()
	for index := 0; index < userCounter; index++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			sendersList[index], err = mnemonic_wallet.New().WalletGenerator(derivation, emptyKey)
			assert.NoError(t, err)

			recipientsList[index], err = mnemonic_wallet.New().WalletGenerator(derivation, emptyKey)
			assert.NoError(t, err)

			var txs []*intMaxTypes.Transfer
			for nonceIndex := 0; nonceIndex < 2; nonceIndex++ {
				salt := new(intMaxTypes.PoseidonHashOut)
				salt.Elements[0] = *new(ffg.Element).SetUint64(0)
				salt.Elements[1] = *new(ffg.Element).SetUint64(0)
				salt.Elements[2] = *new(ffg.Element).SetUint64(0)
				salt.Elements[3] = *new(ffg.Element).SetUint64(uint64(nonceIndex))
				tx := intMaxTypes.Transfer{
					TokenIndex: 0,
					Amount:     amount,
					Salt:       salt,
				}
				if nonceIndex == 0 {
					var publicKey *intMaxAcc.PublicKey
					publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(recipientsList[index].IntMaxWalletAddress)
					assert.NoError(t, err)

					var gaAddr *intMaxTypes.GenericAddress
					gaAddr, err = intMaxTypes.NewINTMAXAddress(publicKey.ToAddress().Bytes())
					assert.NoError(t, err)

					tx.Recipient = gaAddr
				} else {
					const emptyETHAddr = "0x0000000000000000000000000000000000000000"
					addr := common.HexToAddress(recipientsList[index].WalletAddress.String())
					assert.NotEqual(t, addr.String(), emptyETHAddr)

					var gaAddr *intMaxTypes.GenericAddress
					gaAddr, err = intMaxTypes.NewEthereumAddress(addr.Bytes())
					assert.NoError(t, err)

					tx.Recipient = gaAddr
				}
				txs = append(txs, &tx)
			}
			hashTrList := make([][]byte, len(txs))
			for key := range txs {
				hashTrList[key] = txs[key].Hash().Marshal()
			}

			rw := &worker.ReceiverWorker{
				Sender:       sendersList[index].IntMaxWalletAddress,
				TransferHash: hexutil.Encode(keccak256.Hash(hashTrList...)),
				TransferData: txs,
			}

			receiversListForWorker = append(receiversListForWorker, rw)

			err = w.Receiver(rw)
			assert.NoError(t, err)
		}(index)
	}

	for {
		if userCounterCheck <= 0 {
			break
		}
		<-time.After(10 * time.Second)
	}

	if cancel != nil {
		cancel()
	}

	wg.Wait()
}
