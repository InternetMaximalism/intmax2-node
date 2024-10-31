package post_backup_user_state_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	bps "intmax2-node/internal/balance_prover_service"
	intMaxTypes "intmax2-node/internal/types"
	postBackupUserState "intmax2-node/internal/use_cases/post_backup_user_state"
	"intmax2-node/pkg/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	ucPostBackupUserState "intmax2-node/pkg/use_cases/post_backup_user_state"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCasePostBackupUserState(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)
	db := NewMockSQLDriverApp(ctrl)

	uc := ucPostBackupUserState.New(cfg, log, db)

	const (
		path1 = "../../../"
		path2 = "./"
	)

	dir := path1
	if _, err := os.ReadFile(dir + cfg.APP.PEMPathCACert); err != nil {
		dir = path2
	}

	balanceProof, err := intMaxTypes.MakeSamplePlonky2Proof(dir)
	assert.NoError(t, err)
	b64 := balanceProof.ProofBase64String()
	bp, err := intMaxTypes.NewCompressedPlonky2ProofFromBase64String(b64)
	assert.NoError(t, err)
	byteBP, err := bp.MarshalJSON()
	assert.NoError(t, err)
	pi, err := new(bps.BalancePublicInputs).FromPublicInputs(bp.PublicInputs)
	assert.NoError(t, err)

	cases := []struct {
		desc    string
		input   *postBackupUserState.UCPostBackupUserState
		prepare func(*postBackupUserState.UCPostBackupUserState)
		err     error
	}{
		{
			desc: fmt.Sprintf("Error: %s", ucPostBackupUserState.ErrUCPostBackupUserStateInputEmpty),
			err:  ucPostBackupUserState.ErrUCPostBackupUserStateInputEmpty,
		},
		{
			desc: fmt.Sprintf("Error: %s", errorsDB.ErrNotUnique),
			input: &postBackupUserState.UCPostBackupUserState{
				ID:                 uuid.New().String(),
				UserAddress:        uuid.New().String(),
				BalanceProof:       b64,
				EncryptedUserState: uuid.New().String(),
				AuthSignature:      uuid.New().String(),
				BlockNumber:        1,
				CreatedAt:          time.Now().UTC(),
			},
			prepare: func(in *postBackupUserState.UCPostBackupUserState) {
				db.EXPECT().CreateBackupUserState(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mDBApp.UserState{
					ID:                 in.ID,
					UserAddress:        in.UserAddress,
					EncryptedUserState: in.EncryptedUserState,
					AuthSignature:      in.AuthSignature,
					BlockNumber:        in.BlockNumber,
					CreatedAt:          in.CreatedAt,
					UpdatedAt:          in.CreatedAt,
				}, nil)
				db.EXPECT().CreateBalanceProof(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorsDB.ErrNotUnique)
			},
			err: errorsDB.ErrNotUnique,
		},
		{
			desc: "success",
			input: &postBackupUserState.UCPostBackupUserState{
				ID:                 uuid.New().String(),
				UserAddress:        uuid.New().String(),
				BalanceProof:       b64,
				EncryptedUserState: uuid.New().String(),
				AuthSignature:      uuid.New().String(),
				BlockNumber:        1,
				CreatedAt:          time.Now().UTC(),
			},
			prepare: func(in *postBackupUserState.UCPostBackupUserState) {
				db.EXPECT().CreateBackupUserState(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mDBApp.UserState{
					ID:                 in.ID,
					UserAddress:        in.UserAddress,
					EncryptedUserState: in.EncryptedUserState,
					AuthSignature:      in.AuthSignature,
					BlockNumber:        in.BlockNumber,
					CreatedAt:          in.CreatedAt,
					UpdatedAt:          in.CreatedAt,
				}, nil)
				db.EXPECT().CreateBalanceProof(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mDBApp.BalanceProof{
					ID:                     in.ID,
					UserStateID:            in.ID,
					UserAddress:            in.UserAddress,
					BlockNumber:            in.BlockNumber,
					PrivateStateCommitment: pi.PrivateCommitment.String(),
					BalanceProof:           byteBP,
					CreatedAt:              in.CreatedAt,
					UpdatedAt:              in.CreatedAt,
				}, nil)
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].input)
			}

			ctx := context.Background()
			var input *postBackupUserState.UCPostBackupUserStateInput
			if cases[i].input != nil {
				input = &postBackupUserState.UCPostBackupUserStateInput{
					UserAddress:        cases[i].input.UserAddress,
					BalanceProof:       cases[i].input.BalanceProof,
					EncryptedUserState: cases[i].input.EncryptedUserState,
					AuthSignature:      cases[i].input.AuthSignature,
					BlockNumber:        cases[i].input.BlockNumber,
				}
			}

			var result *postBackupUserState.UCPostBackupUserState
			result, err = uc.Do(ctx, input)
			if cases[i].err != nil {
				assert.True(t, errors.Is(err, cases[i].err))
			} else {
				assert.NoError(t, err)
				if cases[i].input != nil {
					assert.Equal(t, result, cases[i].input)
				}
			}
		})
	}
}
