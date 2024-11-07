package block_validity_prover_block_validity_public_inputs_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/block_validity_prover"
	blockValidityProverBlockValidityPublicInputs "intmax2-node/internal/use_cases/block_validity_prover_block_validity_public_inputs"
	"intmax2-node/pkg/logger"
	ucBlockValidityProverBlockValidityPublicInputs "intmax2-node/pkg/use_cases/block_validity_prover_block_validity_public_inputs"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUCBlockValidityProverBlockValidityPublicInputs(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)
	bvs := NewMockBlockValidityService(ctrl)

	uc := ucBlockValidityProverBlockValidityPublicInputs.New(cfg, log, bvs)

	cases := []struct {
		desc    string
		input   *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputsInput
		result  *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs
		prepare func(error)
		err     error
	}{
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBlockValidityPublicInputs.ErrUCBlockValidityProverBlockValidityPublicInputsInputEmpty,
			),
			prepare: func(_ error) {},
			err:     ucBlockValidityProverBlockValidityPublicInputs.ErrUCBlockValidityProverBlockValidityPublicInputsInputEmpty,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				errors.New("fake"),
			),
			input: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputsInput{},
			prepare: func(err error) {
				bvs.EXPECT().ValidityPublicInputsByBlockNumber(gomock.Any()).Return(
					nil,
					nil,
					err,
				)
			},
			err: errors.New("fake"),
		},
		{
			desc: "success",
			input: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputsInput{
				BlockNumber: 1,
			},
			result: &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputs{
				ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{},
				Sender:               make([]block_validity_prover.SenderLeaf, 0),
			},
			prepare: func(_ error) {
				bvs.EXPECT().ValidityPublicInputsByBlockNumber(gomock.Any()).Return(
					&block_validity_prover.ValidityPublicInputs{},
					make([]block_validity_prover.SenderLeaf, 0),
					nil,
				)
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].err)
			}

			ctx := context.Background()
			var input *blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputsInput
			if cases[i].input != nil {
				input = &blockValidityProverBlockValidityPublicInputs.UCBlockValidityProverBlockValidityPublicInputsInput{
					BlockNumber: cases[i].input.BlockNumber,
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
