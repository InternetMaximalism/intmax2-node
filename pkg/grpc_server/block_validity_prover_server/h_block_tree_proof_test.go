package block_validity_prover_server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	blockTreeProofByRootAndLeafBlockNumbers "intmax2-node/internal/use_cases/block_tree_proof_by_root_and_leaf_block_numbers"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dimiro1/health"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestBlockTreeProof(t *testing.T) {
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

	uc := mocks.NewMockUseCaseBlockTreeProofByRootAndLeafBlockNumbers(ctrl)
	_ = uc

	cases := []struct {
		desc            string
		rootBlockNumber string
		leafBlockNumber string
		info            *blockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbers
		prepare         func(info *blockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbers)
		success         bool
		message         string
		wantStatus      int
	}{
		{
			desc:       "empty rootBlockNumber, leafBlockNumber",
			success:    false,
			message:    `type mismatch, parameter: root_block_number, error: strconv.ParseInt: parsing "": invalid syntax`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:            "empty leafBlockNumber",
			rootBlockNumber: "1",
			success:         false,
			message:         `type mismatch, parameter: leaf_block_number, error: strconv.ParseInt: parsing "": invalid syntax`,
			wantStatus:      http.StatusBadRequest,
		},
		{
			desc:            `error with rootBlockNumber = leafBlockNumber = 0.5`,
			rootBlockNumber: "0.5",
			leafBlockNumber: "0.5",
			success:         false,
			message:         `type mismatch, parameter: root_block_number, error: strconv.ParseInt: parsing "0.5": invalid syntax`,
			wantStatus:      http.StatusBadRequest,
		},
		{
			desc:            `error with leafBlockNumber = 0.5`,
			rootBlockNumber: "1",
			leafBlockNumber: "0.5",
			success:         false,
			message:         `type mismatch, parameter: leaf_block_number, error: strconv.ParseInt: parsing "0.5": invalid syntax`,
			wantStatus:      http.StatusBadRequest,
		},
		{
			desc:            "(-1) => leafBlockNumber: must not be less than zero; rootBlockNumber: must not be less than one.",
			rootBlockNumber: "-1",
			leafBlockNumber: "-1",
			success:         false,
			message:         `leafBlockNumber: must not be less than zero; rootBlockNumber: must not be less than one.`,
			wantStatus:      http.StatusBadRequest,
		},
		{
			desc:            "(0) => rootBlockNumber: must not be less than one.",
			rootBlockNumber: "0",
			leafBlockNumber: "0",
			success:         false,
			message:         `rootBlockNumber: must not be less than one.`,
			wantStatus:      http.StatusBadRequest,
		},
		// TODO: added test for use_case blockTreeProofByRootAndLeafBlockNumbers after fixed BlockValidityService.DepositTreeProof()
		{
			desc:            "success",
			rootBlockNumber: "1",
			leafBlockNumber: "1",
			success:         true,
			info: &blockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbers{
				MerkleProof: &blockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbersMerkleProof{
					Siblings: []string{
						uuid.New().String(),
					},
				},
				RootHash: uuid.New().String(),
			},
			prepare: func(info *blockTreeProofByRootAndLeafBlockNumbers.UCBlockTreeProofByRootAndLeafBlockNumbers) {
				cmd.EXPECT().
					BlockTreeProofByRootAndLeafBlockNumbers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uc)
				uc.EXPECT().Do(gomock.Any(), gomock.Any()).Return(info, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].info)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf(
					"http://"+gwServer.Addr+"/v1/block-merkle-tree/%s/%s",
					cases[i].rootBlockNumber,
					cases[i].leafBlockNumber,
				),
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
