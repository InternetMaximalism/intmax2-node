package server_test

import (
	"context"
	"intmax2-node/configs"
	healthCheck "intmax2-node/internal/use_cases/health_check"
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

func TestHandlerHealthCheck(t *testing.T) {
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
	hcTestImpl := newHcTest()
	const hcName = "test"
	hc.AddChecker(hcName, hcTestImpl)

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

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc)
	defer grpcServerStop()

	uc := mocks.NewMockUseCaseHealthCheck(ctrl)

	cases := []struct {
		desc       string
		prepare    func(want bool)
		success    bool
		wantStatus int
	}{
		{
			desc: "Success equal false",
			prepare: func(want bool) {
				cmd.EXPECT().HealthCheck(gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any()).Return(&healthCheck.HealthCheck{
					Success: want,
				})
			},
			success:    false,
			wantStatus: http.StatusOK,
		},
		{
			desc: "Success equal true",
			prepare: func(want bool) {
				cmd.EXPECT().HealthCheck(gomock.Any()).Return(uc)
				uc.EXPECT().Do(gomock.Any()).Return(&healthCheck.HealthCheck{
					Success: want,
				})
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].success)
			}

			hcTestImpl.IsOK(cases[i].success)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://"+gwServer.Addr+"/v1", http.NoBody)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
		})
	}
}

type hcTest interface {
	Check(ctx context.Context) health.Health
	IsOK(ok bool)
}

type hcTestStruct struct {
	ok bool
}

func newHcTest() hcTest {
	return &hcTestStruct{}
}

func (hc *hcTestStruct) Check(_ context.Context) (res health.Health) {
	res.Down()
	if hc.ok {
		res.Up()
	}

	return res
}

func (hc *hcTestStruct) IsOK(ok bool) {
	hc.ok = ok
}
