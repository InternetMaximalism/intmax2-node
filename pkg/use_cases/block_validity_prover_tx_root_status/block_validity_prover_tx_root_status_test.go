package block_validity_prover_tx_root_status_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/block_validity_prover"
	intMaxTypes "intmax2-node/internal/types"
	blockValidityProverTxRootStatus "intmax2-node/internal/use_cases/block_validity_prover_tx_root_status"
	"intmax2-node/pkg/logger"
	ucBlockValidityProverTxRootStatus "intmax2-node/pkg/use_cases/block_validity_prover_tx_root_status"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUCBlockValidityProverTxRootStatus(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	log := logger.New(cfg.LOG.Level, cfg.LOG.TimeFormat, cfg.LOG.JSON, cfg.LOG.IsLogLine)
	bvs := NewMockBlockValidityService(ctrl)

	uc := ucBlockValidityProverTxRootStatus.New(cfg, log, bvs)

	var (
		uint32value1 = uint32(1)
		amountValue  = new(big.Int).SetUint64(123)
		g2Affine     = bn254.G2Affine{}
		bufG2Affine  = g2Affine.RawBytes()
		acc          = &accounts.PublicKey{}

		hashValid = common.Hash{}
		sender    = common.Address{}
	)

	sender.SetBytes(amountValue.Bytes())
	hashValid.SetBytes(amountValue.Bytes())
	_, err := g2Affine.SetBytes(bufG2Affine[:])
	assert.NoError(t, err)
	acc, err = acc.SetBigInt(amountValue)
	assert.NoError(t, err)

	ucResult := map[string]*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus{
		hashValid.String(): {
			IsRegistrationBlock: true,
			TxTreeRoot:          hashValid,
			PrevBlockHash:       hashValid,
			BlockNumber:         uint32value1,
			DepositRoot:         hashValid,
			SignatureHash:       hashValid,
			MessagePoint:        &g2Affine,
			AggregatedPublicKey: acc,
			AggregatedSignature: &g2Affine,
			Senders: []intMaxTypes.Sender{
				{
					PublicKey: acc,
					AccountID: uint64(uint32value1),
					IsSigned:  true,
				},
			},
		},
	}

	cases := []struct {
		desc    string
		input   *blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatusInput
		result  map[string]*blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatus
		prepare func(error)
		err     error
	}{
		{
			desc: fmt.Sprintf(
				"Error: %s",
				ucBlockValidityProverTxRootStatus.ErrUCBlockValidityProverTxRootStatusInputEmpty,
			),
			prepare: func(_ error) {},
			err:     ucBlockValidityProverTxRootStatus.ErrUCBlockValidityProverTxRootStatusInputEmpty,
		},
		{
			desc: fmt.Sprintf(
				"Error: %s",
				errors.New("fake"),
			),
			input: &blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatusInput{},
			prepare: func(err error) {
				bvs.EXPECT().AuxInfoListFromBlockContentByTxRoot(gomock.Any()).Return(
					nil,
					err,
				)
			},
			err: errors.New("fake"),
		},
		{
			desc: "success",
			input: &blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatusInput{
				ConvertTxRoot: []common.Hash{hashValid},
			},
			result: ucResult,
			prepare: func(_ error) {
				bvs.EXPECT().AuxInfoListFromBlockContentByTxRoot(gomock.Any()).Return(map[common.Hash]*block_validity_prover.AuxInfo{
					hashValid: {
						BlockContent: &intMaxTypes.BlockContent{
							IsRegistrationBlock: ucResult[hashValid.String()].IsRegistrationBlock,
							SenderType:          intMaxTypes.PublicKeySenderType,
							Senders:             ucResult[hashValid.String()].Senders,
							TxTreeRoot:          ucResult[hashValid.String()].TxTreeRoot,
							AggregatedSignature: ucResult[hashValid.String()].AggregatedSignature,
							AggregatedPublicKey: ucResult[hashValid.String()].AggregatedPublicKey,
							MessagePoint:        ucResult[hashValid.String()].MessagePoint,
						},
						PostedBlock: &block_post_service.PostedBlock{
							PrevBlockHash: ucResult[hashValid.String()].PrevBlockHash,
							BlockNumber:   ucResult[hashValid.String()].BlockNumber,
							DepositRoot:   ucResult[hashValid.String()].DepositRoot,
							SignatureHash: ucResult[hashValid.String()].SignatureHash,
						},
					},
				}, nil)
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare(cases[i].err)
			}

			ctx := context.Background()
			var input *blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatusInput
			if cases[i].input != nil {
				input = &blockValidityProverTxRootStatus.UCBlockValidityProverTxRootStatusInput{
					ConvertTxRoot: cases[i].input.ConvertTxRoot,
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
