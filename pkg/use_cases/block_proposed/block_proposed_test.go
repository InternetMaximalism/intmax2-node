package block_proposed_test

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTree "intmax2-node/internal/tree"
	"intmax2-node/internal/use_cases/block_proposed"
	"intmax2-node/internal/worker"
	ucBlockProposed "intmax2-node/pkg/use_cases/block_proposed"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCaseBlockProposed(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := ucBlockProposed.New()

	const (
		mnemonic   = "gown situate miss skill figure rain smoke grief giraffe perfect milk gospel casino open mimic egg grace canoe erode skull drip open luggage next"
		mnPassword = ""
		derivation = "m/44'/60'/0'/0/0"

		txHashKey        = "0x22a09569aeffa766a1c0d8d5dd9d3fb3e5b4567700b8cbac3b4eceedeacee793"
		intMaxAddressKey = "0x1c6f2045ddc7fde4f0ff37ac47b2726ed2e6e9fe8ea3d3d6971403cece12306d"
	)

	w, err := mnemonic_wallet.New().WalletFromMnemonic(mnemonic, mnPassword, derivation)
	assert.NoError(t, err)
	assert.Equal(t, w.IntMaxWalletAddress, intMaxAddressKey)

	publicKey, err := intMaxAcc.NewPublicKeyFromAddressHex(intMaxAddressKey)
	assert.NoError(t, err)

	txTree := worker.TxTree{
		TxTreeHash: new(intMaxTree.PoseidonHashOut),
		SenderTransfers: []*worker.SenderTransfers{
			{
				TxTreeLeafHash: new(intMaxTree.PoseidonHashOut),
			},
		},
	}

	cases := []struct {
		desc  string
		input *block_proposed.UCBlockProposedInput
		err   error
	}{
		{
			desc: fmt.Sprintf("Error: %s", ucBlockProposed.ErrUCInputEmpty),
			err:  ucBlockProposed.ErrUCInputEmpty,
		},
		{
			desc: "Success",
			input: &block_proposed.UCBlockProposedInput{
				DecodeSender: publicKey,
				TxHash:       txHashKey,
				TxTree:       &txTree,
			},
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			ctx := context.Background()
			resp, err := uc.Do(ctx, cases[i].input)
			if cases[i].err != nil {
				assert.Nil(t, resp)
				assert.True(t, errors.Is(err, cases[i].err))
			} else {
				assert.NotNil(t, resp)
				assert.NoError(t, err)
			}

			if cases[i].input != nil {
				assert.NotNil(t, resp)
				assert.Equal(t, resp.TxRoot, txTree.TxTreeHash.String())
				assert.Len(t, resp.TxTreeMerkleProof, 1)
				assert.Equal(t, resp.TxTreeMerkleProof[0], txTree.SenderTransfers[0].TxTreeLeafHash.String())
			}
		})
	}
}
