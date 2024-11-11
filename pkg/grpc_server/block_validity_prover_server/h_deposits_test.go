package block_validity_prover_server_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxTree "intmax2-node/internal/tree"
	blockValidityProverDeposits "intmax2-node/internal/use_cases/block_validity_prover_deposits"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/logger"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerDeposits(t *testing.T) {
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

	uc := mocks.NewMockUseCaseBlockValidityProverDeposits(ctrl)

	var (
		uint32value1  = uint32(1)
		amountValue   = new(big.Int).SetUint64(123)
		sender        = common.Address{}
		hashValid     = common.Hash{}
		hashesInvalid []string
	)
	for i := 0; i < cfg.BlockValidityProver.BlockValidityProverMaxValueOfInputDepositsInRequest+1; i++ {
		hashesInvalid = append(hashesInvalid, uuid.New().String())
	}

	sender.SetBytes(amountValue.Bytes())
	hashValid.SetBytes(amountValue.Bytes())

	cases := []struct {
		desc       string
		body       string
		info       []*blockValidityProverDeposits.UCBlockValidityProverDeposits
		prepare    func([]*blockValidityProverDeposits.UCBlockValidityProverDeposits)
		success    bool
		message    string
		wantStatus int
	}{
		{
			desc:       "depositHashes: cannot be blank.",
			message:    "depositHashes: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "depositHashes: cannot be blank.",
			body:       `{"depositHashes":[]}`,
			message:    "depositHashes: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "depositHashes: (0: must be a valid value.).",
			body:       `{"depositHashes":[""]}`,
			message:    "depositHashes: (0: must be a valid value.).",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "depositHashes: (0: must be a valid value.).",
			body:       fmt.Sprintf(`{"depositHashes":[%q]}`, hashesInvalid[0]),
			message:    "depositHashes: (0: must be a valid value.).",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: fmt.Sprintf(
				"depositHashes: the length must be between 1 and %d.",
				cfg.BlockValidityProver.BlockValidityProverMaxValueOfInputDepositsInRequest,
			),
			body: fmt.Sprintf(`{"depositHashes":["%s"]}`, strings.Join(hashesInvalid, `","`)),
			message: fmt.Sprintf(
				"depositHashes: the length must be between 1 and %d.",
				cfg.BlockValidityProver.BlockValidityProverMaxValueOfInputDepositsInRequest,
			),
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "success with UC error `fake`",
			body: fmt.Sprintf(`{"depositHashes":[%q]}`, hashValid.String()),
			prepare: func(_ []*blockValidityProverDeposits.UCBlockValidityProverDeposits) {
				cmd.EXPECT().BlockValidityProverDeposits(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(nil, errors.New("fake"))
			},
			wantStatus: http.StatusOK,
		},
		{
			desc: "success without UC error",
			body: fmt.Sprintf(`{"depositHashes":[%q]}`, hashValid.String()),
			info: []*blockValidityProverDeposits.UCBlockValidityProverDeposits{
				{
					DepositId:      uint32value1,
					DepositHash:    common.Hash{},
					DepositIndex:   &uint32value1,
					BlockNumber:    &uint32value1,
					IsSynchronized: true,
					DepositLeaf: &intMaxTree.DepositLeaf{
						RecipientSaltHash: common.Hash{},
						TokenIndex:        uint32value1,
						Amount:            amountValue,
					},
					Sender: sender.String(),
				},
			},
			prepare: func(input []*blockValidityProverDeposits.UCBlockValidityProverDeposits) {
				cmd.EXPECT().BlockValidityProverDeposits(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(input, nil)
			},
			success:    true,
			wantStatus: http.StatusOK,
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
				"http://"+gwServer.Addr+"/v1/deposits",
				body,
			)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, strings.TrimSpace(gjson.Get(w.Body.String(), "message").String()))
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
			if cases[i].info != nil {
				for key := range cases[i].info {
					assert.Equal(t,
						int64(cases[i].info[key].DepositId),
						gjson.Get(w.Body.String(), fmt.Sprintf("data.deposits.%d.depositId", key)).Int(),
					)
					assert.Equal(t,
						cases[i].info[key].DepositHash.String(),
						gjson.Get(w.Body.String(), fmt.Sprintf("data.deposits.%d.depositHash", key)).String(),
					)
					if assert.NotNil(t, cases[i].info[key].DepositIndex) {
						assert.Equal(t,
							int64(*cases[i].info[key].DepositIndex),
							gjson.Get(w.Body.String(), fmt.Sprintf("data.deposits.%d.depositIndex", key)).Int(),
						)
					}
					if assert.NotNil(t, cases[i].info[key].BlockNumber) {
						assert.Equal(t,
							int64(*cases[i].info[key].BlockNumber),
							gjson.Get(w.Body.String(), fmt.Sprintf("data.deposits.%d.blockNumber", key)).Int(),
						)
					}
					assert.Equal(t,
						cases[i].info[key].IsSynchronized,
						gjson.Get(w.Body.String(), fmt.Sprintf("data.deposits.%d.isSynchronized", key)).Bool(),
					)
					assert.Equal(t,
						cases[i].info[key].Sender,
						gjson.Get(w.Body.String(), fmt.Sprintf("data.deposits.%d.from", key)).String(),
					)
					if assert.NotNil(t, cases[i].info[key].DepositLeaf) {
						assert.Equal(t,
							cases[i].info[key].DepositLeaf.Amount.String(),
							gjson.Get(
								w.Body.String(),
								fmt.Sprintf("data.deposits.%d.depositLeaf.amount", key),
							).String(),
						)
						assert.Equal(t,
							int64(cases[i].info[key].DepositLeaf.TokenIndex),
							gjson.Get(
								w.Body.String(),
								fmt.Sprintf("data.deposits.%d.depositLeaf.tokenIndex", key),
							).Int(),
						)
						assert.Equal(t,
							common.Hash(cases[i].info[key].DepositLeaf.RecipientSaltHash).String(),
							gjson.Get(
								w.Body.String(),
								fmt.Sprintf("data.deposits.%d.depositLeaf.recipientSaltHash", key),
							).String(),
						)
					}
				}
			}
		})
	}
}
