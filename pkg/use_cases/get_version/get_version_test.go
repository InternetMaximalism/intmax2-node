package get_version_test

import (
	"context"
	"intmax2-node/configs"
	intGetVersion "intmax2-node/internal/use_cases/get_version"
	"intmax2-node/internal/use_cases/mocks"
	"intmax2-node/pkg/use_cases/get_version"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	const intValue4 = 3
	assert.NoError(t, configs.LoadDotEnv(intValue4))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("uc_get_version", func(t *testing.T) {
		info := &intGetVersion.Version{
			Version:   uuid.New().String(),
			BuildTime: uuid.New().String(),
		}

		ctx := context.TODO()

		gv := get_version.New(info.Version, info.BuildTime)
		assert.Equal(t, info.Version, gv.Do(ctx).Version)
		assert.Equal(t, info.BuildTime, gv.Do(ctx).BuildTime)

		pkgUC := mocks.NewMockUseCaseGetVersion(ctrl)
		pkgUC.EXPECT().Do(ctx).Return(info)
		assert.Equal(t, info.Version, pkgUC.Do(ctx).Version)
		pkgUC.EXPECT().Do(ctx).Return(info)
		assert.Equal(t, info.BuildTime, pkgUC.Do(ctx).BuildTime)
	})
}
