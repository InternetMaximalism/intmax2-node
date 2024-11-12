package block_validity_prover_server_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
	blockValidityProverTxRootStatus "intmax2-node/internal/use_cases/block_validity_prover_tx_root_status"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/logger"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerTxRootStatus(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)

	dbApp := NewMockSQLDriverApp(ctrl)
	hc := health.NewHandler()
	sb := NewMockServiceBlockchain(ctrl)
	bvs := NewMockBlockValidityService(ctrl)

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

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc, sb, bvs)
	defer grpcServerStop()

	uc := mocks.NewMockUseCaseBlockValidityProverTxRootStatus(ctrl)

	var (
		uint32value1      = uint32(1)
		amountValue       = new(big.Int).SetUint64(123)
		invalidHashAsUUID = uuid.New().String()
		g2Affine          = bn254.G2Affine{}
		bufG2Affine       = g2Affine.RawBytes()
		acc               = &accounts.PublicKey{}
		zeroTxRoot        = "0x0000000000000000000000000000000000000000000000000000000000000000"

		hashesInvalid []string
		hashValid     = common.Hash{}
		sender        = common.Address{}
	)

	for i := 0; i < cfg.BlockValidityProver.BlockValidityProverMaxValueOfInputTxRootInRequest+1; i++ {
		hashesInvalid = append(hashesInvalid, uuid.New().String())
	}

	sender.SetBytes(amountValue.Bytes())
	hashValid.SetBytes(amountValue.Bytes())
	_, err := g2Affine.SetBytes(bufG2Affine[:])
	assert.NoError(t, err)
	acc, err = acc.SetBigInt(amountValue)
	assert.NoError(t, err)

	ucResult := map[string]*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus{
		hashValid.String(): {
			IsRegistrationBlock: true,
			TxTreeRoot:          hashValid,
			PrevBlockHash:       hashValid,
			BlockNumber:         uint32value1,
			DepositRoot:         hashValid,
			SignatureHash:       hashValid,
			MessagePoint:        &g2Affine,
			AggregatedPublicKey: acc,
			AggregatedSignature: &g2Affine,
			Senders: []intMaxTypes.Sender{
				{
					PublicKey: acc,
					AccountID: uint64(uint32value1),
					IsSigned:  true,
				},
			},
		},
	}

	cases := []struct {
		desc          string
		body          string
		info          []*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus
		prepare       func([]*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus)
		success       bool
		message       string
		successTxRoot string
		errorsKey     string
		errorsMessage string
		wantStatus    int
	}{
		{
			desc:       "txRoots: cannot be blank.",
			message:    "txRoots: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "txRoots: cannot be blank.",
			body:       `{"txRoots":[]}`,
			message:    "txRoots: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: fmt.Sprintf(
				"txRoots: the length must be between 1 and %d.",
				cfg.BlockValidityProver.BlockValidityProverMaxValueOfInputTxRootInRequest,
			),
			body: fmt.Sprintf(`{"txRoots":["%s"]}`, strings.Join(hashesInvalid, `","`)),
			message: fmt.Sprintf(
				"txRoots: the length must be between 1 and %d.",
				cfg.BlockValidityProver.BlockValidityProverMaxValueOfInputTxRootInRequest,
			),
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:          "txRoot not existing",
			body:          `{"txRoots":[""]}`,
			success:       true,
			errorsKey:     "",
			errorsMessage: `txRoot not existing`,
			wantStatus:    http.StatusOK,
		},
		{
			desc:          "txRoot not existing",
			body:          fmt.Sprintf(`{"txRoots":[%q]}`, invalidHashAsUUID),
			success:       true,
			errorsKey:     invalidHashAsUUID,
			errorsMessage: `txRoot not existing`,
			wantStatus:    http.StatusOK,
		},
		{
			desc:          "txRoot not existing",
			body:          fmt.Sprintf(`{"txRoots":[%q]}`, zeroTxRoot),
			success:       true,
			errorsKey:     zeroTxRoot,
			errorsMessage: `txRoot not existing`,
			wantStatus:    http.StatusOK,
		},
		{
			desc:          "txRoot not existing",
			body:          fmt.Sprintf(`{"txRoots":[%q]}`, cfg.BlockValidityProver.BlockValidityProverInvalidTxRootInRequest[0]),
			success:       true,
			errorsKey:     cfg.BlockValidityProver.BlockValidityProverInvalidTxRootInRequest[0],
			errorsMessage: `txRoot not existing`,
			wantStatus:    http.StatusOK,
		},
		{
			desc: "txRoot not existing",
			body: fmt.Sprintf(`{"txRoots":[%q]}`, hashValid),
			prepare: func(_ []*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus) {
				cmd.EXPECT().BlockValidityProverTxRootStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			success:       true,
			errorsKey:     hashValid.String(),
			errorsMessage: `txRoot not existing`,
			wantStatus:    http.StatusOK,
		},
		{
			desc: "success with UC error `fake`",
			body: fmt.Sprintf(`{"txRoots":[%q]}`, hashValid),
			prepare: func(_ []*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus) {
				cmd.EXPECT().BlockValidityProverTxRootStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, errors.New("fake"))
			},
			wantStatus: http.StatusOK,
		},
		{
			desc: "success with not empty `data.blocks` and empty `data.errors`",
			body: fmt.Sprintf(`{"txRoots":[%q]}`, hashValid),
			info: []*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus{
				{
					TxTreeRoot: hashValid,
				},
			},
			prepare: func(list []*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus) {
				cmd.EXPECT().BlockValidityProverTxRootStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(ucResult, nil)
			},
			successTxRoot: hashValid.String(),
			success:       true,
			wantStatus:    http.StatusOK,
		},
		{
			desc: "success with not empty `data.blocks` and not empty `data.errors`",
			body: fmt.Sprintf(`{"txRoots":["%s"]}`, strings.Join([]string{hashValid.String(), zeroTxRoot}, `","`)),
			info: []*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus{
				{
					TxTreeRoot: hashValid,
				},
			},
			prepare: func(list []*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus) {
				cmd.EXPECT().BlockValidityProverTxRootStatus(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(ucResult, nil)
			},
			successTxRoot: hashValid.String(),
			errorsKey:     zeroTxRoot,
			errorsMessage: `txRoot not existing`,
			success:       true,
			wantStatus:    http.StatusOK,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].info)
			}

			var body io.Reader
			body = http.NoBody
			if bd := strings.TrimSpace(cases[i].body); bd != "" {
				t.Log("-------- input body", bd)
				body = strings.NewReader(bd)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodPost,
				"http://"+gwServer.Addr+"/v1/tx-root/status",
				body,
			)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, strings.TrimSpace(gjson.Get(w.Body.String(), "message").String()))
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
			if len(gjson.Get(w.Body.String(), "data.blocks").Array()) > 0 {
				assert.Equal(t, cases[i].successTxRoot, gjson.Get(w.Body.String(), fmt.Sprintf("data.blocks.0.txRoot")).String())
			}
			if len(gjson.Get(w.Body.String(), "data.errors").Array()) > 0 {
				assert.Equal(t, cases[i].errorsKey, gjson.Get(w.Body.String(), fmt.Sprintf("data.errors.0.txRoot")).String())
				assert.Equal(t, cases[i].errorsMessage, gjson.Get(w.Body.String(), fmt.Sprintf("data.errors.0.message")).String())
			}
		})
	}
}
