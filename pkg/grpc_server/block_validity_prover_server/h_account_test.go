package block_validity_prover_server_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	blockValidityProverAccount "intmax2-node/internal/use_cases/block_validity_prover_account"
	"intmax2-node/pkg/logger"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	ucBlockValidityProverAccount "intmax2-node/pkg/use_cases/block_validity_prover_account"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dimiro1/health"
	"github.com/google/uuid"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestAccount(t *testing.T) {
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

	cases := []struct {
		desc       string
		address    string
		info       *blockValidityProverAccount.UCBlockValidityProverAccount
		prepare    func(address string, info *blockValidityProverAccount.UCBlockValidityProverAccount)
		success    bool
		message    string
		wantStatus int
	}{
		{
			desc:       `address: cannot be blank.`,
			message:    `address: cannot be blank.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       `address: must be a valid value.`,
			address:    uuid.New().String(),
			message:    `address: must be a valid value.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       `address: must be a valid value.`,
			address:    "0x0000000000000000000000000000000000000000000000000000000000000000",
			message:    `address: must be a valid value.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:    `Internal server error`,
			address: "0x0000000000000000000000000000000000000000000000000000000000000001",
			prepare: func(address string, _ *blockValidityProverAccount.UCBlockValidityProverAccount) {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("fake"))
			},
			message:    `Internal server error`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			desc:    fmt.Sprintf("UC Error: %s", errorsDB.ErrNotFound),
			address: "0x0000000000000000000000000000000000000000000000000000000000000001",
			prepare: func(address string, _ *blockValidityProverAccount.UCBlockValidityProverAccount) {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(errorsDB.ErrNotFound)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
		{
			desc:    fmt.Sprintf("UC Error: %s", ucBlockValidityProverAccount.ErrNewAddressFromHexFail),
			address: "0x0000000000000000000000000000000000000000000000000000000000000001",
			prepare: func(address string, _ *blockValidityProverAccount.UCBlockValidityProverAccount) {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(ucBlockValidityProverAccount.ErrNewAddressFromHexFail)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
		{
			desc:    fmt.Sprintf("UC Error: %s", ucBlockValidityProverAccount.ErrPublicKeyFromIntMaxAccFail),
			address: "0x0000000000000000000000000000000000000000000000000000000000000001",
			prepare: func(address string, _ *blockValidityProverAccount.UCBlockValidityProverAccount) {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(ucBlockValidityProverAccount.ErrPublicKeyFromIntMaxAccFail)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
		{
			desc:    "success",
			address: "0x0000000000000000000000000000000000000000000000000000000000000001",
			info:    &blockValidityProverAccount.UCBlockValidityProverAccount{AccountID: new(uint256.Int).SetUint64(uint64(1))},
			prepare: func(address string, input *blockValidityProverAccount.UCBlockValidityProverAccount) {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, in interface{}, _ func(_ interface{}, _ interface{}) error) error {
						v, ok := in.(*blockValidityProverAccount.UCBlockValidityProverAccount)
						assert.True(t, ok)
						v.AccountID = input.AccountID

						return nil
					},
				)
			},
			success:    true,
			wantStatus: http.StatusOK,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].address, cases[i].info)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("http://"+gwServer.Addr+"/v1/account/%s", cases[i].address),
				http.NoBody,
			)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, strings.TrimSpace(gjson.Get(w.Body.String(), "message").String()))
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
			if gjson.Get(w.Body.String(), "data.isRegistered").Bool() {
				assert.Equal(t, int64(cases[i].info.AccountID.Uint64()), gjson.Get(w.Body.String(), "data.accountId").Int())
			} else {
				assert.Equal(t, int64(0), gjson.Get(w.Body.String(), "data.accountId").Int())
			}
		})
	}
}
