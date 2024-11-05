package block_validity_prover_account_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	blockValidityProverAccount "intmax2-node/internal/use_cases/block_validity_prover_account"
	"intmax2-node/pkg/logger"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	ucBlockValidityProverAccount "intmax2-node/pkg/use_cases/block_validity_prover_account"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUCBlockValidityProverAccount(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)
	db := NewMockSQLDriverApp(ctrl)

	uc := ucBlockValidityProverAccount.New(cfg, log, db)

	cases := []struct {
		desc    string
		input   *blockValidityProverAccount.UCBlockValidityProverAccountInput
		result  *blockValidityProverAccount.UCBlockValidityProverAccount
		prepare func(*blockValidityProverAccount.UCBlockValidityProverAccount)
		err     error
	}{
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverAccount.ErrUCBlockValidityProverAccountInputEmpty,
			),
			err: ucBlockValidityProverAccount.ErrUCBlockValidityProverAccountInputEmpty,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverAccount.ErrNewAddressFromHexFail,
			),
			input: &blockValidityProverAccount.UCBlockValidityProverAccountInput{
				Address: uuid.New().String(),
			},
			err: ucBlockValidityProverAccount.ErrNewAddressFromHexFail,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				errorsDB.ErrNotFound,
			),
			input: &blockValidityProverAccount.UCBlockValidityProverAccountInput{
				Address: "0x0000000000000000000000000000000000000000000000000000000000000000",
			},
			prepare: func(_ *blockValidityProverAccount.UCBlockValidityProverAccount) {
				db.EXPECT().SenderByAddress(gomock.Any()).Return(nil, errorsDB.ErrNotFound)
			},
			err: errorsDB.ErrNotFound,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				errorsDB.ErrNotFound,
			),
			input: &blockValidityProverAccount.UCBlockValidityProverAccountInput{
				Address: "0x0000000000000000000000000000000000000000000000000000000000000000",
			},
			prepare: func(input *blockValidityProverAccount.UCBlockValidityProverAccount) {
				db.EXPECT().SenderByAddress(gomock.Any()).Return(&mDBApp.Sender{
					ID:        uuid.New().String(),
					Address:   "0000000000000000000000000000000000000000000000000000000000000000",
					PublicKey: "0000000000000000000000000000000000000000000000000000000000000000",
					CreatedAt: time.Now().UTC(),
				}, nil)
				db.EXPECT().AccountBySenderID(gomock.Any()).Return(nil, errorsDB.ErrNotFound)
			},
			err: errorsDB.ErrNotFound,
		},
		{
			desc: "success",
			input: &blockValidityProverAccount.UCBlockValidityProverAccountInput{
				Address: "0x0000000000000000000000000000000000000000000000000000000000000000",
			},
			result: &blockValidityProverAccount.UCBlockValidityProverAccount{
				AccountID: new(uint256.Int).SetUint64(123),
			},
			prepare: func(result *blockValidityProverAccount.UCBlockValidityProverAccount) {
				sender := mDBApp.Sender{
					ID:        uuid.New().String(),
					Address:   "0000000000000000000000000000000000000000000000000000000000000000",
					PublicKey: "0000000000000000000000000000000000000000000000000000000000000000",
					CreatedAt: time.Now().UTC(),
				}
				db.EXPECT().SenderByAddress(gomock.Any()).Return(&sender, nil)
				db.EXPECT().AccountBySenderID(gomock.Any()).Return(&mDBApp.Account{
					ID:        uuid.New().String(),
					AccountID: result.AccountID,
					SenderID:  sender.ID,
					CreatedAt: time.Now().UTC(),
				}, nil)
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].result)
			}

			ctx := context.Background()
			var input *blockValidityProverAccount.UCBlockValidityProverAccountInput
			if cases[i].input != nil {
				input = &blockValidityProverAccount.UCBlockValidityProverAccountInput{
					Address: cases[i].input.Address,
				}
			}

			result, err := uc.Do(ctx, input)
			if cases[i].err != nil {
				assert.True(t, errors.Is(err, cases[i].err))
			} else {
				assert.NoError(t, err)
				if cases[i].result != nil {
					assert.Equal(t, result, cases[i].result)
				}
			}
		})
	}
}
