package block_validity_prover_server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	depositTreeProofByDepositIndex "intmax2-node/internal/use_cases/deposit_tree_proof_by_deposit_index"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/dimiro1/health"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerDepositTreeProof(t *testing.T) {
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

	ucDepositTreeProofByDepositIndex := mocks.NewMockUseCaseDepositTreeProofByDepositIndex(ctrl)

	cases := []struct {
		desc         string
		depositIndex string
		info         *depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex
		prepare      func(depositIndex uint32, info *depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex)
		success      bool
		message      string
		wantStatus   int
	}{
		{
			desc:       "empty depositIndex",
			success:    false,
			message:    `type mismatch, parameter: depositIndex, error: strconv.ParseInt: parsing "": invalid syntax`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:         `error with depositIndex = 0.5`,
			depositIndex: "0.5",
			success:      false,
			message:      `type mismatch, parameter: depositIndex, error: strconv.ParseInt: parsing "0.5": invalid syntax`,
			wantStatus:   http.StatusBadRequest,
		},
		{
			desc:         "depositIndex: must not be less than one. (-1)",
			depositIndex: "-1",
			success:      false,
			message:      `depositIndex: must not be less than one.`,
			wantStatus:   http.StatusBadRequest,
		},
		{
			desc:         "depositIndex: must not be less than one. (0)",
			depositIndex: "0",
			success:      false,
			message:      `depositIndex: must not be less than one.`,
			wantStatus:   http.StatusBadRequest,
		},
		// TODO: remove this test after fixed BlockValidityService.DepositTreeProof()
		{
			desc:         fmt.Sprintf("error with status %d", http.StatusInternalServerError),
			depositIndex: "1",
			success:      false,
			prepare: func(depositIndex uint32, _ *depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex) {
				cmd.EXPECT().
					DepositTreeProofByDepositIndex(gomock.Any(), gomock.Any(), gomock.Any())
			},
			message:    `(node) panic triggered: runtime error: invalid memory address or nil pointer dereference`,
			wantStatus: http.StatusInternalServerError,
		},
		// TODO: edited this test after fixed BlockValidityService.DepositTreeProof()
		{
			desc:         "success",
			depositIndex: "1",
			success:      true,
			info: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex{
				MerkleProof: depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexMerkleProof{
					Siblings: []string{
						uuid.New().String(),
					},
				},
				RootHash: uuid.New().String(),
			},
			prepare: func(depositIndex uint32, info *depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex) {
				cmd.EXPECT().
					DepositTreeProofByDepositIndex(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(ucDepositTreeProofByDepositIndex)
				ucDepositTreeProofByDepositIndex.EXPECT().Do(gomock.Any(), gomock.Any()).Return(info, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				di, err := strconv.Atoi(cases[i].depositIndex)
				assert.NoError(t, err)
				cases[i].prepare(uint32(di), cases[i].info)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://"+gwServer.Addr+"/v1/deposit-tree-proof/%s", cases[i].depositIndex),
				http.NoBody,
			)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			if cases[i].wantStatus == http.StatusOK {
				bi, err := json.Marshal(&cases[i].info)
				assert.NoError(t, err)
				assert.Equal(
					t,
					string(bi),
					strings.ReplaceAll(gjson.Get(w.Body.String(), "data").String(), " ", ""),
				)
			}

			assert.Equal(t, cases[i].message, strings.TrimSpace(gjson.Get(w.Body.String(), "message").String()))
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
		})
	}
}
