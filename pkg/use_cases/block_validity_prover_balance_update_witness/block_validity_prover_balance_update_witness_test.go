package block_validity_prover_balance_update_witness_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	blockValidityProverBalanceUpdateWitness "intmax2-node/internal/use_cases/block_validity_prover_balance_update_witness"
	"intmax2-node/pkg/logger"
	ucBlockValidityProverBalanceUpdateWitness "intmax2-node/pkg/use_cases/block_validity_prover_balance_update_witness"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUCBlockValidityProverBalanceUpdateWitness(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)
	bvs := NewMockBlockValidityService(ctrl)

	uc := ucBlockValidityProverBalanceUpdateWitness.New(cfg, log, bvs)

	randPH1, err := goldenposeidon.NewPoseidonHashOut().SetRandom()
	assert.NoError(t, err)
	randPH2, err := goldenposeidon.NewPoseidonHashOut().SetRandom()
	assert.NoError(t, err)
	randPH3, err := goldenposeidon.NewPoseidonHashOut().SetRandom()
	assert.NoError(t, err)
	siblings := []*goldenposeidon.PoseidonHashOut{
		randPH1,
		randPH2,
		randPH3,
	}

	accountMembershipProof := intMaxTree.IndexedMembershipProof{
		IsIncluded: true,
		LeafProof: intMaxTree.IndexedMerkleProof{
			Siblings: siblings,
		},
		LeafIndex: 10,
		Leaf: intMaxTree.IndexedMerkleLeaf{
			Key:       new(big.Int).SetUint64(123),
			Value:     11,
			NextIndex: 12,
			NextKey:   new(big.Int).SetUint64(321),
		},
	}

	validityProof := uuid.New().String()

	cases := []struct {
		desc    string
		input   *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput
		ucResp  *block_validity_prover.UpdateWitness
		result  *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness
		prepare func(*block_validity_prover.UpdateWitness)
		err     error
	}{
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBalanceUpdateWitness.ErrUCBlockValidityProverBalanceUpdateWitnessInputEmpty,
			),
			prepare: func(_ *block_validity_prover.UpdateWitness) {},
			err:     ucBlockValidityProverBalanceUpdateWitness.ErrUCBlockValidityProverBalanceUpdateWitnessInputEmpty,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBalanceUpdateWitness.ErrNewAddressFromHexFail,
			),
			input: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
				User: uuid.New().String(),
			},
			prepare: func(_ *block_validity_prover.UpdateWitness) {},
			err:     ucBlockValidityProverBalanceUpdateWitness.ErrNewAddressFromHexFail,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBalanceUpdateWitness.ErrPublicKeyFromIntMaxAccFail,
			),
			input: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
				User: "0x0000000000000000000000000000000000000000000000000000000000000000",
			},
			prepare: func(_ *block_validity_prover.UpdateWitness) {},
			err:     ucBlockValidityProverBalanceUpdateWitness.ErrPublicKeyFromIntMaxAccFail,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberLessThenTargetBlockNumber,
			),
			input: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
				User: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
			prepare: func(_ *block_validity_prover.UpdateWitness) {
				bvs.EXPECT().FetchUpdateWitness(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, block_validity_prover.ErrRootBlockNumberLessThenLeafBlockNumber)
			},
			err: ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberLessThenTargetBlockNumber,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberInvalid,
			),
			input: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
				User: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
			prepare: func(_ *block_validity_prover.UpdateWitness) {
				bvs.EXPECT().FetchUpdateWitness(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, block_validity_prover.ErrCurrentBlockNumberNotFound)
			},
			err: ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberInvalid,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberInvalid,
			),
			input: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
				User: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
			prepare: func(_ *block_validity_prover.UpdateWitness) {
				bvs.EXPECT().FetchUpdateWitness(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.Join(block_validity_prover.ErrBlockTreeProofFail, block_validity_prover.ErrRootBlockNumberNotFound))
			},
			err: ucBlockValidityProverBalanceUpdateWitness.ErrCurrentBlockNumberInvalid,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBalanceUpdateWitness.ErrTargetBlockNumberInvalid,
			),
			input: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
				User: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
			prepare: func(_ *block_validity_prover.UpdateWitness) {
				bvs.EXPECT().FetchUpdateWitness(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.Join(block_validity_prover.ErrBlockTreeProofFail, block_validity_prover.ErrLeafBlockNumberNotFound))
			},
			err: ucBlockValidityProverBalanceUpdateWitness.ErrTargetBlockNumberInvalid,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverBalanceUpdateWitness.ErrFetchUpdateWitnessFail,
			),
			input: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
				User: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
			prepare: func(_ *block_validity_prover.UpdateWitness) {
				bvs.EXPECT().FetchUpdateWitness(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, errors.New("fake"))
			},
			err: ucBlockValidityProverBalanceUpdateWitness.ErrFetchUpdateWitnessFail,
		},
		{
			desc: "success",
			input: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
				User: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
			ucResp: &block_validity_prover.UpdateWitness{
				ValidityProof: validityProof,
				BlockMerkleProof: intMaxTree.BlockHashMerkleProof{
					Siblings: siblings,
				},
				AccountMembershipProof: &accountMembershipProof,
			},
			result: &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitness{
				IsPrevAccountTree: false,
				ValidityProof:     validityProof,
				BlockMerkleProof: intMaxTree.BlockHashMerkleProof{
					Siblings: siblings,
				},
				AccountMembershipProof: &accountMembershipProof,
			},
			prepare: func(updW *block_validity_prover.UpdateWitness) {
				bvs.EXPECT().FetchUpdateWitness(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(&block_validity_prover.UpdateWitness{
					ValidityProof:          updW.ValidityProof,
					BlockMerkleProof:       updW.BlockMerkleProof,
					AccountMembershipProof: updW.AccountMembershipProof,
				}, nil)
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].ucResp)
			}

			ctx := context.Background()
			var input *blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput
			if cases[i].input != nil {
				input = &blockValidityProverBalanceUpdateWitness.UCBlockValidityProverBalanceUpdateWitnessInput{
					User:               cases[i].input.User,
					CurrentBlockNumber: cases[i].input.CurrentBlockNumber,
					TargetBlockNumber:  cases[i].input.TargetBlockNumber,
					IsPrevAccountTree:  cases[i].input.IsPrevAccountTree,
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
