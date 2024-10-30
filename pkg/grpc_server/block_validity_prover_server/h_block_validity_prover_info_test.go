package block_validity_prover_server_test

import (
	"context"
	"intmax2-node/configs"
	blockValidityProverInfo "intmax2-node/internal/use_cases/block_validity_prover_info"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dimiro1/health"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerBlockValidityProverInfo(t *testing.T) {
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

	uc := mocks.NewMockUseCaseBlockValidityProverInfo(ctrl)

	cases := []struct {
		desc       string
		info       *blockValidityProverInfo.UCBlockValidityProverInfo
		prepare    func(version *blockValidityProverInfo.UCBlockValidityProverInfo)
		success    bool
		wantStatus int
	}{
		{
			desc: "Success",
			info: &blockValidityProverInfo.UCBlockValidityProverInfo{
				DepositIndex: 123,
				BlockNumber:  321,
			},
			prepare: func(info *blockValidityProverInfo.UCBlockValidityProverInfo) {
				cmd.EXPECT().BlockValidityProverInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any()).Return(info, nil)
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

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://"+gwServer.Addr+"/v1/info", http.NoBody)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].info.BlockNumber, gjson.Get(w.Body.String(), "data.blockNumber").Int())
			assert.Equal(t, cases[i].info.DepositIndex, gjson.Get(w.Body.String(), "data.depositIndex").Int())
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
		})
	}
}
