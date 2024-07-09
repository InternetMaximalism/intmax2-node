package transaction_test

import (
	"context"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/use_cases/transaction"
	ucTransaction "intmax2-node/pkg/use_cases/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCaseTransaction(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	dbApp := NewMockSQLDriverApp(ctrl)
	worker := NewMockWorker(ctrl)

	uc := ucTransaction.New(cfg, dbApp, worker)

	cases := []struct {
		desc    string
		input   *transaction.UCTransactionInput
		prepare func()
		err     error
	}{
		{
			desc: "Success",
			prepare: func() {
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare()
			}

			ctx := context.TODO()
			if cases[i].err != nil {
				assert.True(t, errors.Is(uc.Do(ctx, cases[i].input), cases[i].err))
			} else {
				assert.NoError(t, uc.Do(ctx, cases[i].input))
			}
		})
	}
}
