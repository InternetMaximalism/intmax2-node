package deposit_tree_proof_by_deposit_index_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxTree "intmax2-node/internal/tree"
	depositTreeProofByDepositIndex "intmax2-node/internal/use_cases/deposit_tree_proof_by_deposit_index"
	"intmax2-node/pkg/logger"
	ucDepositTreeProofByDepositIndex "intmax2-node/pkg/use_cases/deposit_tree_proof_by_deposit_index"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDepositTreeProofByDepositIndex(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)
	bvs := NewMockBlockValidityService(ctrl)

	uc := ucDepositTreeProofByDepositIndex.New(cfg, log, bvs)

	type testType struct {
		DepositMerkleProof *intMaxTree.KeccakMerkleProof
		DepositTreeRoot    common.Hash
	}

	hash := common.BigToHash(new(big.Int).SetUint64(123))
	nodeHash := [32]byte{}
	copy(nodeHash[:], hash[:])

	cases := []struct {
		desc    string
		input   *depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput
		ucResp  *testType
		result  *depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex
		prepare func(*testType, error)
		err     error
	}{
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucDepositTreeProofByDepositIndex.ErrUCDepositTreeProofByDepositIndexInputEmpty,
			),
			prepare: func(_ *testType, _ error) {},
			err:     ucDepositTreeProofByDepositIndex.ErrUCDepositTreeProofByDepositIndexInputEmpty,
		},
		{
			desc: "use LatestDepositTreeProofByBlockNumber with fake error",
			input: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput{
				DepositIndex: 0,
				BlockNumber:  0,
			},
			prepare: func(_ *testType, err error) {
				bvs.EXPECT().LatestDepositTreeProofByBlockNumber(gomock.Any()).Return(
					nil,
					common.Hash{},
					err,
				)
			},
			err: errors.New("fake"),
		},
		{
			desc: "use LatestDepositTreeProofByBlockNumber",
			input: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput{
				DepositIndex: 0,
				BlockNumber:  0,
			},
			ucResp: &testType{
				DepositMerkleProof: &intMaxTree.KeccakMerkleProof{Siblings: [][32]byte{
					nodeHash,
				}},
				DepositTreeRoot: hash,
			},
			result: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex{
				MerkleProof: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexMerkleProof{
					Siblings: []string{
						common.BytesToHash(nodeHash[:]).String(),
					},
				},
				RootHash: hash.String(),
			},
			prepare: func(ucResp *testType, _ error) {
				bvs.EXPECT().LatestDepositTreeProofByBlockNumber(gomock.Any()).Return(
					ucResp.DepositMerkleProof,
					ucResp.DepositTreeRoot,
					nil,
				)
			},
		},
		{
			desc: "use DepositTreeProof with fake error",
			input: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput{
				DepositIndex: 0,
				BlockNumber:  1,
			},
			prepare: func(_ *testType, err error) {
				bvs.EXPECT().DepositTreeProof(gomock.Any(), gomock.Any()).Return(
					nil,
					common.Hash{},
					err,
				)
			},
			err: errors.New("fake"),
		},
		{
			desc: "use DepositTreeProof",
			input: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput{
				DepositIndex: 0,
				BlockNumber:  1,
			},
			ucResp: &testType{
				DepositMerkleProof: &intMaxTree.KeccakMerkleProof{Siblings: [][32]byte{
					nodeHash,
				}},
				DepositTreeRoot: hash,
			},
			result: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndex{
				MerkleProof: &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexMerkleProof{
					Siblings: []string{
						common.BytesToHash(nodeHash[:]).String(),
					},
				},
				RootHash: hash.String(),
			},
			prepare: func(ucResp *testType, _ error) {
				bvs.EXPECT().DepositTreeProof(gomock.Any(), gomock.Any()).Return(
					ucResp.DepositMerkleProof,
					ucResp.DepositTreeRoot,
					nil,
				)
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil || cases[i].err != nil {
				cases[i].prepare(cases[i].ucResp, cases[i].err)
			}

			ctx := context.Background()
			var input *depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput
			if cases[i].input != nil {
				input = &depositTreeProofByDepositIndex.UCDepositTreeProofByDepositIndexInput{
					DepositIndex: cases[i].input.DepositIndex,
					BlockNumber:  cases[i].input.BlockNumber,
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
