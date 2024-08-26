package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/pow"
	"intmax2-node/internal/tx_transfer_service"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_signature"
	"intmax2-node/internal/worker"
	"intmax2-node/pkg/logger"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerBlockSignature(t *testing.T) {
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

	sampleData := makeSampleData(t)
	sampleDataJson, err := json.Marshal(sampleData)
	if err != nil {
		require.NoError(t, err)
	}

	cmd := NewMockCommands(ctrl)
	//ucBS := mocks.NewMockUseCaseBlockSignature(ctrl)

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc, pwNonce, wrk, sb, storageGPO)
	defer grpcServerStop()

	cases := []struct {
		desc       string
		prepare    func()
		body       string
		success    bool
		message    string
		wantStatus int
	}{
		// validation - start
		{
			desc:       "Empty body",
			prepare:    func() {},
			message:    "enoughBalanceProof: (prevBalanceProof: (proof: cannot be blank; publicInputs: cannot be blank.); transferStepProof: (proof: cannot be blank; publicInputs: cannot be blank.).); sender: cannot be blank; signature: cannot be blank; txHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Empty body",
			prepare:    func() {},
			body:       `{}`,
			message:    "enoughBalanceProof: (prevBalanceProof: (proof: cannot be blank; publicInputs: cannot be blank.); transferStepProof: (proof: cannot be blank; publicInputs: cannot be blank.).); sender: cannot be blank; signature: cannot be blank; txHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Invalid signature",
			prepare:    func() {},
			body:       fmt.Sprintf(`{"signature":%q}`, uuid.New().String()),
			message:    "enoughBalanceProof: (prevBalanceProof: (proof: cannot be blank; publicInputs: cannot be blank.); transferStepProof: (proof: cannot be blank; publicInputs: cannot be blank.).); sender: cannot be blank; signature: must be a valid value; txHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "Invalid sender",
			prepare:    func() {},
			body:       fmt.Sprintf(`{"sender":%q}`, uuid.New().String()),
			message:    "enoughBalanceProof: (prevBalanceProof: (proof: cannot be blank; publicInputs: cannot be blank.); transferStepProof: (proof: cannot be blank; publicInputs: cannot be blank.).); sender: must be a valid value; signature: cannot be blank; txHash: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "Success",
			prepare: func() {
				wrk.EXPECT().TrHash(gomock.Any()).Return(&worker.TransactionHashesWithSenderAndFile{
					Sender: sampleData.Sender,
					TxHash: sampleData.TxHash,
					File:   &os.File{},
				}, nil)
				wrk.EXPECT().TxTreeByAvailableFile(gomock.Any()).Return(sampleData.TxTree, nil)
				fmt.Println("------------ sampleData.TxTree", sampleData.TxTree)
			},
			body:       string(sampleDataJson),
			message:    block_signature.SuccessMsg,
			wantStatus: http.StatusOK,
		},
		// validation - finish
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare()
			}

			var body io.Reader
			body = http.NoBody
			if bd := strings.TrimSpace(cases[i].body); bd != "" {
				t.Log("-------- input body", bd)
				body = strings.NewReader(bd)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://"+gwServer.Addr+"/v1/block/signature", body)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, gjson.Get(w.Body.String(), "message").String())
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
		})
	}
}

func makeSampleData(t *testing.T) *block_signature.UCBlockSignatureInput {
	senderAccount, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(big.NewInt(2))
	assert.NoError(t, err)

	transferTreeRoot, err := tx_transfer_service.MakeSampleTransferTree()
	assert.NoError(t, err)

	txTreeRoot, err := tx_transfer_service.MakeSampleTxTree(&transferTreeRoot, 1)
	assert.NoError(t, err)

	senderPublicKeys := []*intMaxAcc.PublicKey{}
	for i := 0; i < 4; i++ {
		sa, err := intMaxAcc.NewPrivateKeyWithReCalcPubKeyIfPkNegates(big.NewInt(int64(i) + 2))
		assert.NoError(t, err)

		senderPublicKeys = append(senderPublicKeys, sa.Public())
	}

	senderPublicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senderPublicKeys {
		senderPublicKey := sender.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[32*i:32*(i+1)], senderPublicKey[:])
	}

	defaultPublicKey := intMaxAcc.NewDummyPublicKey().Pk.X.Bytes() // Only x coordinate is used
	for i := len(senderPublicKeys); i < intMaxTypes.NumOfSenders; i++ {
		copy(senderPublicKeysBytes[32*i:32*(i+1)], defaultPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)

	var res *block_signature.UCBlockSignatureInput
	res, err = tx_transfer_service.MakePostBlockSignatureRawRequest(
		senderAccount, txTreeRoot, publicKeysHash,
	)
	require.NoError(t, err)

	return res
}
