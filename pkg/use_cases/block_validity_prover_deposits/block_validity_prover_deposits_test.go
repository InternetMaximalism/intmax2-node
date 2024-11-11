package block_validity_prover_deposits_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"
	blockValidityProverDeposits "intmax2-node/internal/use_cases/block_validity_prover_deposits"
	"intmax2-node/pkg/logger"
	ucBlockValidityProverDeposits "intmax2-node/pkg/use_cases/block_validity_prover_deposits"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUCBlockValidityProverDeposits(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)
	bvs := NewMockBlockValidityService(ctrl)

	uc := ucBlockValidityProverDeposits.New(cfg, log, bvs)

	var (
		uint32value1 = uint32(1)
		amountValue  = new(big.Int).SetUint64(123)
		sender       = common.Address{}
		hash         = common.Hash{}
	)

	sender.SetBytes(amountValue.Bytes())
	hash.SetBytes(amountValue.Bytes())

	cases := []struct {
		desc    string
		input   *blockValidityProverDeposits.UCBlockValidityProverDepositsInput
		result  []*blockValidityProverDeposits.UCBlockValidityProverDeposits
		prepare func([]*blockValidityProverDeposits.UCBlockValidityProverDeposits, error)
		err     error
	}{
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverDeposits.ErrUCBlockValidityProverDepositsInputEmpty,
			),
			prepare: func(_ []*blockValidityProverDeposits.UCBlockValidityProverDeposits, _ error) {},
			err:     ucBlockValidityProverDeposits.ErrUCBlockValidityProverDepositsInputEmpty,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				errors.New("fake"),
			),
			input: &blockValidityProverDeposits.UCBlockValidityProverDepositsInput{},
			prepare: func(_ []*blockValidityProverDeposits.UCBlockValidityProverDeposits, err error) {
				bvs.EXPECT().GetDepositsInfoByHash(gomock.Any()).Return(
					nil,
					err,
				)
			},
			err: errors.New("fake"),
		},
		{
			desc: "success",
			input: &blockValidityProverDeposits.UCBlockValidityProverDepositsInput{
				ConvertDepositHashes: []common.Hash{hash},
			},
			result: []*blockValidityProverDeposits.UCBlockValidityProverDeposits{
				{
					DepositId:      uint32value1,
					DepositHash:    hash,
					DepositIndex:   &uint32value1,
					BlockNumber:    &uint32value1,
					IsSynchronized: true,
					DepositLeaf: &intMaxTree.DepositLeaf{
						RecipientSaltHash: hash,
						TokenIndex:        uint32value1,
						Amount:            amountValue,
					},
					Sender: sender.String(),
				},
			},
			prepare: func(input []*blockValidityProverDeposits.UCBlockValidityProverDeposits, _ error) {
				result := make(map[uint32]*block_validity_prover.DepositInfo)
				for key := range input {
					result[input[key].DepositId] = &block_validity_prover.DepositInfo{
						DepositId:      input[key].DepositId,
						DepositHash:    input[key].DepositHash,
						DepositIndex:   input[key].DepositIndex,
						BlockNumber:    input[key].BlockNumber,
						IsSynchronized: input[key].IsSynchronized,
						DepositLeaf:    input[key].DepositLeaf,
						Sender:         input[key].Sender,
					}
				}
				bvs.EXPECT().GetDepositsInfoByHash(gomock.Any()).Return(result, nil)
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].result, cases[i].err)
			}

			ctx := context.Background()
			var input *blockValidityProverDeposits.UCBlockValidityProverDepositsInput
			if cases[i].input != nil {
				input = &blockValidityProverDeposits.UCBlockValidityProverDepositsInput{
					ConvertDepositHashes: cases[i].input.ConvertDepositHashes,
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
