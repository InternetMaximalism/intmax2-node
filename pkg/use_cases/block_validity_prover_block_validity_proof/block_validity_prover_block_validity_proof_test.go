package block_validity_prover_block_validity_proof_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/block_validity_prover"
	blockValidityProverBlockValidityProof "intmax2-node/internal/use_cases/block_validity_prover_block_validity_proof"
	"intmax2-node/pkg/logger"
	ucBlockValidityProverBlockValidityProof "intmax2-node/pkg/use_cases/block_validity_prover_block_validity_proof"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUCBlockValidityProverBlockValidityProof(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)
	bvs := NewMockBlockValidityService(ctrl)

	uc := ucBlockValidityProverBlockValidityProof.New(cfg, log, bvs)

	validityProof := uuid.New().String()

	cases := []struct {
		desc    string
		input   *blockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProofInput
		result  *blockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProof
		prepare func(error)
		err     error
	}{
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBlockValidityProof.ErrUCBlockValidityProverBlockValidityProofInputEmpty,
			),
			prepare: func(_ error) {},
			err:     ucBlockValidityProverBlockValidityProof.ErrUCBlockValidityProverBlockValidityProofInputEmpty,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				errors.New("fake"),
			),
			input: &blockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProofInput{},
			prepare: func(err error) {
				bvs.EXPECT().ValidityProofByBlockNumber(gomock.Any()).Return(
					nil,
					err,
				)
			},
			err: errors.New("fake"),
		},
		{
			desc: "success",
			input: &blockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProofInput{
				BlockNumber: 1,
			},
			result: &blockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProof{
				ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{},
				ValidityProof:        &validityProof,
				Sender:               make([]block_validity_prover.SenderLeaf, 0),
			},
			prepare: func(_ error) {
				bvs.EXPECT().ValidityProofByBlockNumber(gomock.Any()).Return(
					&block_validity_prover.ValidityProof{
						ValidityPublicInputs: &block_validity_prover.ValidityPublicInputs{},
						SenderLeaf:           make([]block_validity_prover.SenderLeaf, 0),
						ValidityProof:        &validityProof,
					},
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
			var input *blockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProofInput
			if cases[i].input != nil {
				input = &blockValidityProverBlockValidityProof.UCBlockValidityProverBlockValidityProofInput{
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
