package store_vault_server_test

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	bps "intmax2-node/internal/balance_prover_service"
	intMaxTypes "intmax2-node/internal/types"
	postBackupUserState "intmax2-node/internal/use_cases/post_backup_user_state"
	"intmax2-node/pkg/logger"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dimiro1/health"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
)

func TestBackupUserStateTest(t *testing.T) {
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

	grpcServerStop, gwServer := Start(cmd, ctx, cfg, log, dbApp, &hc, sb)
	defer grpcServerStop()

	balanceProof, err := intMaxTypes.MakeSamplePlonky2Proof(dir)
	assert.NoError(t, err)
	b64 := balanceProof.ProofBase64String()
	bp, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(b64)
	assert.NoError(t, err)
	_, err = new(bps.BalancePublicInputs).FromPublicInputs(bp.PublicInputs)
	assert.NoError(t, err)

	cases := []struct {
		desc       string
		resp       *postBackupUserState.UCPostBackupUserState
		prepare    func()
		body       string
		success    bool
		successMsg string
		message    string
		wantStatus int
	}{
		{
			desc:       "All Empty",
			message:    `authSignature: cannot be blank; balanceProof: cannot be blank; blockNumber: cannot be blank; encryptedUserState: cannot be blank; userAddress: cannot be blank.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "blockNumber not empty",
			body:       `{"blockNumber":1}`,
			message:    `authSignature: cannot be blank; balanceProof: cannot be blank; encryptedUserState: cannot be blank; userAddress: cannot be blank.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "blockNumber not empty; userAddress not empty",
			body:       `{"blockNumber":1,"userAddress":"userAddress"}`,
			message:    `authSignature: cannot be blank; balanceProof: cannot be blank; encryptedUserState: cannot be blank.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "blockNumber not empty; userAddress not empty; encryptedUserState not empty",
			body:       `{"blockNumber":1,"userAddress":"userAddress","encryptedUserState":"encryptedUserState"}`,
			message:    `authSignature: cannot be blank; balanceProof: cannot be blank.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "blockNumber not empty; userAddress not empty; encryptedUserState not empty; authSignature not empty",
			body:       `{"blockNumber":1,"userAddress":"userAddress","encryptedUserState":"encryptedUserState","authSignature":"authSignature"}`,
			message:    `balanceProof: cannot be blank.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc:       "blockNumber not empty; userAddress not empty; encryptedUserState not empty; authSignature not empty; balanceProof is invalid value",
			body:       `{"blockNumber":1,"userAddress":"userAddress","encryptedUserState":"encryptedUserState","authSignature":"authSignature","balanceProof":"balanceProof"}`,
			message:    `balanceProof: must be a valid value.`,
			wantStatus: http.StatusBadRequest,
		},
		{
			desc: "blockNumber not empty; userAddress not empty; encryptedUserState not empty; authSignature not empty; balanceProof is valid value",
			body: fmt.Sprintf(`{"blockNumber":1,"userAddress":"userAddress","encryptedUserState":"encryptedUserState","authSignature":"authSignature","balanceProof":%q}`, b64),
			resp: &postBackupUserState.UCPostBackupUserState{},
			prepare: func() {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(errorsDB.ErrNotUnique)
			},
			success:    false,
			successMsg: postBackupUserState.AlreadyExistsMsg,
			wantStatus: http.StatusOK,
		},
		{
			desc: "success",
			body: fmt.Sprintf(`{"blockNumber":1,"userAddress":"userAddress","encryptedUserState":"encryptedUserState","authSignature":"authSignature","balanceProof":%q}`, b64),
			resp: &postBackupUserState.UCPostBackupUserState{},
			prepare: func() {
				dbApp.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			success:    true,
			successMsg: postBackupUserState.SuccessMsg,
			wantStatus: http.StatusOK,
		},
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
			r := httptest.NewRequest(http.MethodPost, "http://"+gwServer.Addr+"/v1/backups/user-state", body)

			gwServer.Handler.ServeHTTP(w, r)

			if !assert.Equal(t, cases[i].wantStatus, w.Code) {
				t.Log(w.Body.String())
			}

			assert.Equal(t, cases[i].message, gjson.Get(w.Body.String(), "message").String())
			assert.Equal(t, cases[i].success, gjson.Get(w.Body.String(), "success").Bool())
			if cases[i].resp != nil {
				assert.Equal(t, cases[i].successMsg, gjson.Get(w.Body.String(), "data.message").String())
			}
		})
	}
}
