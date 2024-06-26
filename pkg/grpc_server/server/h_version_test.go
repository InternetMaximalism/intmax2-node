package server_test

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	getVersion "intmax2-node/internal/use_cases/get_version"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dimiro1/health"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestHandlerVersion(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)

	pw := NewMockPoWNonce(ctrl)
	dbApp := NewMockSQLDriverApp(ctrl)
	worker := NewMockWorker(ctrl)
	hc := health.NewHandler()

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

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc, pw, worker)
	defer grpcServerStop()

	getVer := mocks.NewMockUseCaseGetVersion(ctrl)

	cases := []struct {
		desc       string
		info       *getVersion.Version
		prepare    func(version *getVersion.Version)
		wantStatus int
	}{
		{
			desc: "Success",
			info: &getVersion.Version{
				Version:   uuid.New().String(),
				BuildTime: uuid.New().String(),
			},
			prepare: func(info *getVersion.Version) {
				buildvars.Version = info.Version
				buildvars.BuildTime = info.BuildTime
				cmd.EXPECT().GetVersion(gomock.Any(), gomock.Any()).Return(getVer)
				getVer.EXPECT().Do(gomock.Any()).Return(info)
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
			r := httptest.NewRequest(http.MethodGet, "http://"+gwServer.Addr+"/v1/version", http.NoBody)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].info.Version, gjson.Get(w.Body.String(), "version").String())
			assert.Equal(t, cases[i].info.BuildTime, gjson.Get(w.Body.String(), "buildtime").String())
		})
	}
}
