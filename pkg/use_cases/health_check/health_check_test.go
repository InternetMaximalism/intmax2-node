package health_check_test

import (
	"context"
	"intmax2-node/configs"
	healthCheck "intmax2-node/internal/use_cases/health_check"
	"intmax2-node/internal/use_cases/mocks"
	ucHealthCheck "intmax2-node/pkg/use_cases/health_check"
	"testing"

	"github.com/dimiro1/health"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCaseHealthCheck(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pkgUC := mocks.NewMockUseCaseHealthCheck(ctrl)

	hc := health.NewHandler()
	hcTestImpl := newHcTest()
	const hcName = "test"
	hc.AddChecker(hcName, hcTestImpl)

	uc := ucHealthCheck.New(&hc)

	cases := []struct {
		desc    string
		prepare func(hc *healthCheck.HealthCheck)
		success bool
	}{
		{
			desc: "Is down",
			prepare: func(hc *healthCheck.HealthCheck) {
				pkgUC.EXPECT().Do(gomock.Any()).Return(hc)
			},
			success: false,
		},
		{
			desc: "Is up",
			prepare: func(hc *healthCheck.HealthCheck) {
				pkgUC.EXPECT().Do(gomock.Any()).Return(hc)
			},
			success: true,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(&healthCheck.HealthCheck{
					Success: cases[i].success,
				})
			}

			ctx := context.TODO()

			hcTestImpl.IsOK(cases[i].success)
			if cases[i].success {
				assert.True(t, uc.Do(ctx).Success)
			} else {
				assert.False(t, uc.Do(ctx).Success)
			}
			assert.Equal(t, uc.Do(ctx).Success, pkgUC.Do(ctx).Success)
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
