package store_vault_server_test

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	getBackupUserState "intmax2-node/internal/use_cases/get_backup_user_state"
	"intmax2-node/pkg/logger"
	errorsDB "intmax2-node/pkg/sql_db/errors"
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

func TestGetBackupUserState(t *testing.T) {
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

	const (
		emptyKey = ""
		path1    = "../../../"
		path2    = "./"
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

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc, sb)
	defer grpcServerStop()

	cases := []struct {
		desc        string
		userStateID string
		prepare     func()
		success     bool
		successMsg  string
		message     string
		wantStatus  int
	}{
		{
			desc:       "UserStateID empty",
			message:    "userStateId: cannot be blank.",
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:        "UserStateID not found",
			userStateID: uuid.New().String(),
			successMsg:  getBackupUserState.NotFoundMessage,
			prepare: func() {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(errorsDB.ErrNotFound)
			},
			wantStatus: http.StatusOK,
		},
		{
			desc:        "Success",
			userStateID: uuid.New().String(),
			successMsg:  getBackupUserState.SuccessMsg,
			prepare: func() {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare()
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://"+gwServer.Addr+"/v1/backups/user-state/%s", cases[i].userStateID),
				http.NoBody,
			)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, gjson.Get(w.Body.String(), "message").String())
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
			if strings.EqualFold(cases[i].userStateID, emptyKey) {
				assert.Equal(t, cases[i].successMsg, gjson.Get(w.Body.String(), "data.message").String())
			}
		})
	}
}
