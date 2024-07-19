package server_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/pow"
	"intmax2-node/internal/use_cases/block_proposed"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/internal/worker"
	"intmax2-node/pkg/logger"
	ucBlockProposed "intmax2-node/pkg/use_cases/block_proposed"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerBlockProposed(t *testing.T) {
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
	ucBP := mocks.NewMockUseCaseBlockProposed(ctrl)

	const (
		mnemonic   = "gown situate miss skill figure rain smoke grief giraffe perfect milk gospel casino open mimic egg grace canoe erode skull drip open luggage next"
		mnPassword = ""
		derivation = "m/44'/60'/0'/0/0"

		txHashKey        = "0x3098f91cabb2463569a5158ffd0d7cd2420dece6556d41b4eda120a5a937892e"
		intMaxAddressKey = "0x1c6f2045ddc7fde4f0ff37ac47b2726ed2e6e9fe8ea3d3d6971403cece12306d"
	)

	w, err := mnemonic_wallet.New().WalletFromMnemonic(mnemonic, mnPassword, derivation)
	assert.NoError(t, err)
	assert.Equal(t, w.IntMaxWalletAddress, intMaxAddressKey)

	expiration := time.Now().Add(60 * time.Minute)

	var signature string
	{
		pk, err := intMaxAcc.HexToPrivateKey(w.IntMaxPrivateKey)
		assert.NoError(t, err)

		keyPair, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(pk.BigInt())
		assert.NoError(t, err)

		message, err := block_proposed.MakeMessage(txHashKey, w.IntMaxWalletAddress, expiration)
		assert.NoError(t, err)

		sign, err := keyPair.Sign(message)
		assert.NoError(t, err)
		signature = hexutil.Encode(sign.Marshal())
	}

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc, pwNonce, wrk)
	defer grpcServerStop()

	cases := []struct {
		desc       string
		resp       *block_proposed.UCBlockProposed
		prepare    func(*block_proposed.UCBlockProposed)
		body       string
		success    bool
		message    string
		wantStatus int
	}{
		// validation - start
		{
			desc:       "Empty body",
			prepare:    func(_ *block_proposed.UCBlockProposed) {},
			message:    "expiration: must be a valid value; sender: cannot be blank; signature: cannot be blank; txHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Empty body",
			prepare:    func(_ *block_proposed.UCBlockProposed) {},
			body:       `{}`,
			message:    "expiration: must be a valid value; sender: cannot be blank; signature: cannot be blank; txHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Invalid signature",
			prepare:    func(_ *block_proposed.UCBlockProposed) {},
			body:       fmt.Sprintf(`{"signature":%q}`, uuid.New().String()),
			message:    "expiration: must be a valid value; sender: cannot be blank; signature: must be a valid value; txHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Invalid sender",
			prepare:    func(_ *block_proposed.UCBlockProposed) {},
			body:       fmt.Sprintf(`{"sender":%q}`, uuid.New().String()),
			message:    "expiration: must be a valid value; sender: must be a valid value; signature: cannot be blank; txHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Invalid txHash",
			prepare:    func(_ *block_proposed.UCBlockProposed) {},
			body:       fmt.Sprintf(`{"txHash":%q}`, uuid.New().String()),
			message:    "expiration: must be a valid value; sender: cannot be blank; signature: cannot be blank; txHash: must be a valid value.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: fmt.Sprintf("Error: %s", block_proposed.ErrTransactionHashNotFound),
			prepare: func(_ *block_proposed.UCBlockProposed) {
				wrk.EXPECT().TrHash(gomock.Any()).Return(nil, worker.ErrTransactionHashNotFound)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"txHash":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, txHashKey),
			message:    fmt.Sprintf("txHash: %s.", block_proposed.ErrTransactionHashNotFound),
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: fmt.Sprintf("Error: %s", block_proposed.ErrTransactionHashNotFound),
			prepare: func(_ *block_proposed.UCBlockProposed) {
				wrk.EXPECT().TrHash(gomock.Any()).Return(&worker.TransactionHashesWithSenderAndFile{
					Sender: uuid.New().String(),
				}, nil)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"txHash":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, txHashKey),
			message:    fmt.Sprintf("txHash: %s.", block_proposed.ErrTransactionHashNotFound),
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: fmt.Sprintf("Error: %s", block_proposed.ErrTransactionHashNotFound),
			prepare: func(_ *block_proposed.UCBlockProposed) {
				wrk.EXPECT().TrHash(gomock.Any()).Return(&worker.TransactionHashesWithSenderAndFile{
					Sender: intMaxAddressKey,
				}, nil)
				wrk.EXPECT().TxTreeByAvailableFile(gomock.Any()).Return(nil, worker.ErrTxTreeByAvailableFileFail)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"txHash":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, txHashKey),
			message:    fmt.Sprintf("txHash: %s.", block_proposed.ErrTransactionHashNotFound),
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: fmt.Sprintf("Error: %s", block_proposed.ErrTxTreeNotBuild),
			prepare: func(_ *block_proposed.UCBlockProposed) {
				wrk.EXPECT().TrHash(gomock.Any()).Return(&worker.TransactionHashesWithSenderAndFile{
					Sender: intMaxAddressKey,
				}, nil)
				wrk.EXPECT().TxTreeByAvailableFile(gomock.Any()).Return(nil, worker.ErrTxTreeNotFound)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"txHash":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, txHashKey),
			message:    fmt.Sprintf("txHash: %s.", block_proposed.ErrTxTreeNotBuild),
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: fmt.Sprintf("Error: %s", block_proposed.ErrTxTreeSignatureCollectionComplete),
			prepare: func(_ *block_proposed.UCBlockProposed) {
				wrk.EXPECT().TrHash(gomock.Any()).Return(&worker.TransactionHashesWithSenderAndFile{
					Sender: intMaxAddressKey,
				}, nil)
				wrk.EXPECT().TxTreeByAvailableFile(gomock.Any()).Return(nil, worker.ErrTxTreeSignatureCollectionComplete)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"txHash":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, txHashKey),
			message:    fmt.Sprintf("txHash: %s.", block_proposed.ErrTxTreeSignatureCollectionComplete),
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: fmt.Sprintf("Error: %s", block_proposed.ErrValueInvalid),
			prepare: func(_ *block_proposed.UCBlockProposed) {
				wrk.EXPECT().TrHash(gomock.Any()).Return(&worker.TransactionHashesWithSenderAndFile{
					Sender: intMaxAddressKey,
				}, nil)
				wrk.EXPECT().TxTreeByAvailableFile(gomock.Any()).Return(nil, errors.New("fake"))
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"txHash":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, txHashKey),
			message:    fmt.Sprintf("txHash: %s.", block_proposed.ErrValueInvalid),
			wantStatus: http.StatusBadRequest,
		},
		// validation - finish
		// uc error - start
		{
			desc: "Internal server error",
			prepare: func(_ *block_proposed.UCBlockProposed) {
				wrk.EXPECT().TrHash(gomock.Any()).Return(&worker.TransactionHashesWithSenderAndFile{
					Sender: intMaxAddressKey,
				}, nil)
				wrk.EXPECT().TxTreeByAvailableFile(gomock.Any())
				cmd.EXPECT().BlockProposed().Return(ucBP)
				ucBP.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, ucBlockProposed.ErrUCInputEmpty)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"txHash":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, txHashKey),
			message:    "Internal server error",
			wantStatus: http.StatusInternalServerError,
		},
		// uc error - finish
		{
			desc: "Valid request",
			resp: &block_proposed.UCBlockProposed{
				TxRoot:            uuid.New().String(),
				TxTreeMerkleProof: []string{uuid.New().String()},
			},
			prepare: func(resp *block_proposed.UCBlockProposed) {
				wrk.EXPECT().TrHash(gomock.Any()).Return(&worker.TransactionHashesWithSenderAndFile{
					Sender: intMaxAddressKey,
				}, nil)
				wrk.EXPECT().TxTreeByAvailableFile(gomock.Any())
				cmd.EXPECT().BlockProposed().Return(ucBP)
				ucBP.EXPECT().Do(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			body:       fmt.Sprintf(`{"sender":%q,"expiration":%q,"signature":%q,"txHash":%q}`, intMaxAddressKey, expiration.Format(time.RFC3339), signature, txHashKey),
			success:    true,
			wantStatus: http.StatusOK,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].resp)
			}

			var body io.Reader
			body = http.NoBody
			if bd := strings.TrimSpace(cases[i].body); bd != "" {
				t.Log("-------- input body", bd)
				body = strings.NewReader(bd)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://"+gwServer.Addr+"/v1/block/proposed", body)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, gjson.Get(w.Body.String(), "message").String())
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
			if cases[i].resp != nil {
				assert.Equal(t, cases[i].resp.TxRoot, gjson.Get(w.Body.String(), "data.txRoot").String())
				assert.Len(t, cases[i].resp.TxTreeMerkleProof, 1)
				assert.Equal(t, fmt.Sprintf("[%q]", cases[i].resp.TxTreeMerkleProof[0]), gjson.Get(w.Body.String(), "data.txTreeMerkleProof").String())
			}
		})
	}
}
