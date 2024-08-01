package transaction_test

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/pow"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	"intmax2-node/internal/worker"
	ucTransaction "intmax2-node/pkg/use_cases/transaction"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCaseTransaction(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	w := NewMockWorker(ctrl)

	uc := ucTransaction.New(cfg, w)

	cases := []struct {
		desc    string
		input   *transaction.UCTransactionInput
		prepare func()
		err     error
	}{
		{
			desc: fmt.Sprintf("Error: %s", worker.ErrReceiverWorkerEmpty.Error()),
			input: func() *transaction.UCTransactionInput {
				return MakeSampleTxRequest(t, cfg)
			}(),
			prepare: func() {
				w.EXPECT().Receiver(gomock.Any()).Return(worker.ErrReceiverWorkerEmpty)
			},
			err: worker.ErrReceiverWorkerEmpty,
		},
		{
			desc: fmt.Sprintf("Error: %s", worker.ErrReceiverWorkerDuplicate.Error()),
			input: func() *transaction.UCTransactionInput {
				return MakeSampleTxRequest(t, cfg)
			}(),
			prepare: func() {
				w.EXPECT().Receiver(gomock.Any()).Return(worker.ErrReceiverWorkerDuplicate)
			},
			err: worker.ErrReceiverWorkerDuplicate,
		},
		{
			desc: fmt.Sprintf("Error: %s", worker.ErrRegisterReceiverFail.Error()),
			input: func() *transaction.UCTransactionInput {
				return MakeSampleTxRequest(t, cfg)
			}(),
			prepare: func() {
				w.EXPECT().Receiver(gomock.Any()).Return(worker.ErrRegisterReceiverFail)
			},
			err: worker.ErrRegisterReceiverFail,
		},
		{
			desc: "Success",
			input: func() *transaction.UCTransactionInput {
				return MakeSampleTxRequest(t, cfg)
			}(),
			prepare: func() {
				w.EXPECT().Receiver(gomock.Any())
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare()
			}

			ctx := context.Background()
			if cases[i].err != nil {
				assert.True(t, errors.Is(uc.Do(ctx, cases[i].input), cases[i].err))
			} else {
				assert.NoError(t, uc.Do(ctx, cases[i].input))
			}
		})
	}
}

func MakeSampleTxRequest(t *testing.T, cfg *configs.Config) *transaction.UCTransactionInput {
	senderAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	assert.NoError(t, err)
	sender := senderAccount.ToAddress().String()

	recipientAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(4))
	assert.NoError(t, err)
	recipientAddress, err := intMaxTypes.NewINTMAXAddress(recipientAccount.ToAddress().Bytes())
	assert.NoError(t, err)

	salt := new(intMaxTypes.PoseidonHashOut).SetZero()
	zeroHash := new(intMaxTypes.PoseidonHashOut).SetZero()
	transfer := intMaxTypes.Transfer{
		Recipient:  recipientAddress,
		TokenIndex: 1,
		Amount:     big.NewInt(100),
		Salt:       salt,
	}
	transfers := make([]*intMaxTypes.Transfer, 8)
	transfers[0] = &transfer
	transferTree, err := intMaxTree.NewTransferTree(
		7,
		nil,
		zeroHash,
	)
	assert.NoError(t, err)
	transfersHash, _, _ := transferTree.GetCurrentRootCountAndSiblings()
	var nonce uint64 = 1

	expiration := time.Now()

	ctx := context.Background()
	pw := pow.New(cfg.PoW.Difficulty)
	pWorker := pow.NewWorker(cfg.PoW.Workers, pw)
	powNonce, err := pow.NewPoWNonce(pw, pWorker).Nonce(ctx, transfersHash.Marshal())
	assert.NoError(t, err)

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	message := crypto.Keccak256(
		transfersHash.Marshal(), nonceBytes, []byte(powNonce), []byte(sender), []byte(expiration.Format(time.RFC3339)),
	)

	signature, err := senderAccount.Sign(finite_field.BytesToFieldElementSlice(message))
	assert.NoError(t, err)

	return &transaction.UCTransactionInput{
		Sender:        sender,
		DecodeSender:  senderAccount.Public(),
		TransfersHash: transfersHash.String(),
		Nonce:         nonce,
		PowNonce:      powNonce,
		Expiration:    expiration,
		Signature:     hexutil.Encode(signature.Marshal()),
	}
}
