package server_test

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/pow"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/internal/worker"
	"intmax2-node/pkg/logger"
	ucTransaction "intmax2-node/pkg/use_cases/transaction"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerTransaction(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)

	dbApp := NewMockSQLDriverApp(ctrl)
	wrk := NewMockWorker(ctrl)
	hc := health.NewHandler()
	sb := NewMockServiceBlockchain(ctrl)
	storageGPO := NewMockGPOStorage(ctrl)

	pw := pow.New(cfg.PoW.Difficulty)
	pWorker := pow.NewWorker(cfg.PoW.Workers, pw)
	pwNonce := pow.NewPoWNonce(pw, pWorker)

	const (
		path1 = "../../../"
		path2 = "./"
	)

	dir := path1
	if _, err := os.ReadFile(dir + cfg.APP.PEMPathCACert); err != nil {
		dir = path2
	}
	cfg.APP.PEMPathCACert = dir + cfg.APP.PEMPathCACert
	cfg.APP.PEMPathServCert = dir + cfg.APP.PEMPathServCert
	cfg.APP.PEMPathServKey = dir + cfg.APP.PEMPathServKey
	cfg.APP.PEMPAthCACertClient = dir + cfg.APP.PEMPAthCACertClient
	cfg.APP.PEMPathClientCert = dir + cfg.APP.PEMPathClientCert
	cfg.APP.PEMPathClientKey = dir + cfg.APP.PEMPathClientKey

	cmd := NewMockCommands(ctrl)
	ucTr := mocks.NewMockUseCaseTransaction(ctrl)

	const (
		mnemonic   = "gown situate miss skill figure rain smoke grief giraffe perfect milk gospel casino open mimic egg grace canoe erode skull drip open luggage next"
		mnPassword = ""
		derivation = "m/44'/60'/0'/0/0"

		nonce            = 1
		amount           = 10
		trHashKey        = "0x22a09569aeffa766a1c0d8d5dd9d3fb3e5b4567700b8cbac3b4eceedeacee793"
		ethAddressKey    = "0xD7fa191fB4F255f7Af801966819382edDA19E09C"
		intMaxAddressKey = "0x2a0a9871a59d52c3d52f57d0ab4324662f39ce14bd2e7a9e2f4c01212b6bea84"
	)

	w, err := mnemonic_wallet.New().WalletFromMnemonic(mnemonic, mnPassword, derivation)
	assert.NoError(t, err)
	assert.Equal(t, w.IntMaxWalletAddress, intMaxAddressKey)

	expiration := time.Now().Add(60 * time.Minute)

	var (
		signature string
		noncePW   string
	)
	_ = signature
	{
		pk, err := intMaxAcc.HexToPrivateKey(w.IntMaxPrivateKey)
		assert.NoError(t, err)

		keyPair, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(pk.BigInt())
		assert.NoError(t, err)

		transfersHash, err := hexutil.Decode(trHashKey)
		if err != nil {
			assert.NoError(t, err)
		}

		trHash := new(intMaxTypes.PoseidonHashOut)
		err = trHash.Unmarshal(transfersHash)
		assert.NoError(t, err)
		tx, err := intMaxTypes.NewTx(
			trHash,
			nonce,
		)
		assert.NoError(t, err)
		txHash := tx.Hash()
		messageForPow := txHash.Marshal()
		noncePW, err = pwNonce.Nonce(ctx, messageForPow)
		assert.NoError(t, err)

		walletAddress, err := intMaxAcc.NewAddressFromHex(w.IntMaxWalletAddress)
		assert.NoError(t, err)

		message, err := transaction.MakeMessage(transfersHash, nonce, noncePW, walletAddress, expiration)
		assert.NoError(t, err)

		sign, err := keyPair.Sign(message)
		assert.NoError(t, err)
		signature = hexutil.Encode(sign.Marshal())
	}

	var currTx *intMaxTypes.Tx
	{
		salt := new(intMaxTypes.PoseidonHashOut)
		salt.Elements[0] = *new(ffg.Element).SetUint64(uint64(1))
		salt.Elements[1] = *new(ffg.Element).SetUint64(uint64(2))
		salt.Elements[2] = *new(ffg.Element).SetUint64(uint64(3))
		salt.Elements[3] = *new(ffg.Element).SetUint64(uint64(4))

		var gaAddr *intMaxTypes.GenericAddress
		gaAddr, err = intMaxTypes.NewEthereumAddress(common.HexToAddress(ethAddressKey).Bytes())
		assert.NoError(t, err)

		tx := intMaxTypes.Transfer{
			Recipient:  gaAddr,
			TokenIndex: 0,
			Amount:     new(big.Int).SetInt64(int64(amount)),
			Salt:       salt,
		}

		transferTree, err := intMaxTree.NewTransferTree(
			intMaxTree.TX_TREE_HEIGHT,
			[]*intMaxTypes.Transfer{&tx},
			intMaxGP.NewPoseidonHashOut(),
		)
		assert.NoError(t, err)

		transferRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()

		currTx, err = intMaxTypes.NewTx(
			&transferRoot,
			nonce,
		)
		assert.NoError(t, err)
	}

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc, pwNonce, wrk, sb, storageGPO)
	defer grpcServerStop()

	cases := []struct {
		desc       string
		prepare    func()
		body       string
		success    bool
		message    string
		dataMsg    string
		wantStatus int
	}{
		{
			desc:       "Empty body",
			prepare:    func() {},
			message:    "expiration: must be a valid value; nonce: cannot be blank; powNonce: cannot be blank; sender: cannot be blank; signature: cannot be blank; transfersHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Invalid transfersHash",
			prepare:    func() {},
			body:       fmt.Sprintf(`{"transfersHash":%q}`, uuid.New().String()),
			message:    "expiration: must be a valid value; nonce: cannot be blank; powNonce: cannot be blank; sender: cannot be blank; signature: cannot be blank; transfersHash: must be a valid value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Valid nonce",
			prepare:    func() {},
			body:       fmt.Sprintf(`{"transfersHash":"0x","nonce":%s}`, "10000000000000000000"),
			message:    "expiration: must be a valid value; powNonce: cannot be blank; sender: cannot be blank; signature: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Invalid powNonce",
			prepare:    func() {},
			body:       fmt.Sprintf(`{"transfersHash":"0x","nonce":%d,"powNonce":%q}`, 0, uuid.New().String()),
			message:    "expiration: must be a valid value; nonce: cannot be blank; powNonce: failed to unmarshal transfers hash; sender: cannot be blank; signature: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Invalid signature",
			prepare:    func() {},
			body:       fmt.Sprintf(`{"transfersHash":"0x","nonce":%d,"powNonce":%q,"signature":%q}`, 0, uuid.New().String(), uuid.New().String()),
			message:    "expiration: must be a valid value; nonce: cannot be blank; powNonce: failed to unmarshal transfers hash; sender: cannot be blank; signature: must be a valid value.",
			wantStatus: http.StatusBadRequest,
		},
		// uc error - start
		{
			desc: fmt.Sprintf("Error: %s", transaction.NotUniqueMsg),
			prepare: func() {
				cmd.EXPECT().Transaction(gomock.Any(), gomock.Any(), gomock.Any()).Return(ucTr)
				ucTr.EXPECT().Do(gomock.Any(), gomock.Any()).Return(worker.ErrReceiverWorkerDuplicate)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"transfersHash":"0x22a09569aeffa766a1c0d8d5dd9d3fb3e5b4567700b8cbac3b4eceedeacee793","nonce":%d,"powNonce":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, nonce, noncePW),
			message:    transaction.NotUniqueMsg,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "Internal server error",
			prepare: func() {
				cmd.EXPECT().Transaction(gomock.Any(), gomock.Any(), gomock.Any()).Return(ucTr)
				ucTr.EXPECT().Do(gomock.Any(), gomock.Any()).Return(ucTransaction.ErrUCInputEmpty)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"transfersHash":"0x22a09569aeffa766a1c0d8d5dd9d3fb3e5b4567700b8cbac3b4eceedeacee793","nonce":%d,"powNonce":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, nonce, noncePW),
			message:    "Internal server error",
			wantStatus: http.StatusInternalServerError,
		},
		{
			desc: "Internal server error",
			prepare: func() {
				cmd.EXPECT().Transaction(gomock.Any(), gomock.Any(), gomock.Any()).Return(ucTr)
				ucTr.EXPECT().Do(gomock.Any(), gomock.Any()).Return(ucTransaction.ErrTransferWorkerReceiverFail)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"transfersHash":"0x22a09569aeffa766a1c0d8d5dd9d3fb3e5b4567700b8cbac3b4eceedeacee793","nonce":%d,"powNonce":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, nonce, noncePW),
			message:    "Internal server error",
			wantStatus: http.StatusInternalServerError,
		},
		// uc error - finish
		// check transfersHash with transferData - start
		{
			desc: "Valid request with transaction to ETHEREUM address",
			prepare: func() {
				cmd.EXPECT().Transaction(gomock.Any(), gomock.Any(), gomock.Any()).Return(ucTr)
				ucTr.EXPECT().Do(gomock.Any(), gomock.Any())
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"transfersHash":"0x22a09569aeffa766a1c0d8d5dd9d3fb3e5b4567700b8cbac3b4eceedeacee793","nonce":%d,"powNonce":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, nonce, noncePW),
			success:    true,
			dataMsg:    transaction.SuccessMsg,
			wantStatus: http.StatusOK,
		},
		// check transfersHash with transferData - finish
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare()
			}

			var body io.Reader
			body = http.NoBody
			if bd := strings.TrimSpace(cases[i].body); bd != "" {
				if cases[i].wantStatus == http.StatusOK {
					t.Log("-------- currTx", currTx.Hash().String())
				}
				t.Log("-------- input body", bd)
				body = strings.NewReader(bd)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://"+gwServer.Addr+"/v1/transaction", body)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, gjson.Get(w.Body.String(), "message").String())
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
			assert.Equal(t, cases[i].dataMsg, gjson.Get(w.Body.String(), "data.message").String())
		})
	}
}
