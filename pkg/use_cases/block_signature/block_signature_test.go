package block_signature_test

import (
	"context"
	"encoding/hex"
	"errors"
	"intmax2-node/configs"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/mnemonic_wallet"
	"testing"

	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	blockSignature "intmax2-node/internal/use_cases/block_signature"
	ucBlockSignature "intmax2-node/pkg/use_cases/block_signature"

	"github.com/stretchr/testify/assert"
)

func TestUseCaseTransaction(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	uc := ucBlockSignature.New()

	const (
		mnPassword = ""
		derivation = "m/44'/60'/0'/0/0"
	)

	senderAccount, err := mnemonic_wallet.New().WalletGenerator(
		derivation, mnPassword,
	)
	assert.NoError(t, err)
	sender := senderAccount.IntMaxWalletAddress

	wrongSenderAccount, err := mnemonic_wallet.New().WalletGenerator(
		derivation, mnPassword,
	)
	assert.NoError(t, err)
	wrongSender := wrongSenderAccount.IntMaxWalletAddress

	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	transferTree, err := intMaxTree.NewTransferTree(6, nil, zeroTransfer.Hash())
	assert.NoError(t, err)
	transferTreeRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()
	zeroTx, err := intMaxTypes.NewTx(&transferTreeRoot, 0)
	assert.NoError(t, err)
	txTree, err := intMaxTree.NewTxTree(7, nil, zeroTx.Hash())
	assert.NoError(t, err)
	txTreeRoot, _, _ := txTree.GetCurrentRootCountAndSiblings()
	txHash := txTreeRoot.Marshal()

	signer, err := intMaxAcc.NewPrivateKeyFromString(senderAccount.IntMaxPrivateKey)
	assert.NoError(t, err)
	flattenMessage := finite_field.BytesToFieldElementSlice(txHash)
	signature, err := signer.Sign(flattenMessage)
	assert.NoError(t, err)

	enoughBalanceProof := &blockSignature.EnoughBalanceProofInput{
		PrevBalanceProof: &blockSignature.Plonky2Proof{
			PublicInputs: []string{
				"2726224824249046055", "14025881618846813748", "5361314524880173070", "2912484915938769214",
			},
			Proof: "0x99396b28",
		},
		TransferStepProof: &blockSignature.Plonky2Proof{
			PublicInputs: []string{
				"2726224824249046055", "14025881618846813748", "5361314524880173070", "2912484915938769214",
			},
			Proof: "0x99396b28",
		},
	} // dummy
	wrongEnoughBalanceProof := new(blockSignature.EnoughBalanceProofInput).Set(enoughBalanceProof)
	wrongEnoughBalanceProof.PrevBalanceProof.PublicInputs = []string{
		"2726224824249046055", "14025881618846813748", "5361314524880173070", "2912484915938769215",
	} // dummy

	cases := []struct {
		desc  string
		input *blockSignature.UCBlockSignatureInput
		err   error
	}{
		{
			desc: "Success",
			input: &blockSignature.UCBlockSignatureInput{
				Sender:             sender,
				TxHash:             hex.EncodeToString(txHash),
				Signature:          hex.EncodeToString(signature.Marshal()),
				EnoughBalanceProof: enoughBalanceProof,
			},
			err: nil,
		},
		{
			desc: "Fail to verify signature",
			input: &blockSignature.UCBlockSignatureInput{
				Sender:             wrongSender,
				TxHash:             hex.EncodeToString(txHash),
				Signature:          hex.EncodeToString(signature.Marshal()),
				EnoughBalanceProof: enoughBalanceProof,
			},
			err: ucBlockSignature.ErrInvalidSignature,
		},
		{
			desc: "Invalid enough balance proof",
			input: &blockSignature.UCBlockSignatureInput{
				Sender:             wrongSender,
				TxHash:             hex.EncodeToString(txHash),
				Signature:          hex.EncodeToString(signature.Marshal()),
				EnoughBalanceProof: wrongEnoughBalanceProof,
			},
			err: ucBlockSignature.ErrInvalidEnoughBalanceProof,
		},
	}

	for i := range cases {
		t.Run(cases[i].desc, func(t *testing.T) {
			ctx := context.Background()
			_, err := uc.Do(ctx, cases[i].input)
			if cases[i].err != nil {
				assert.True(t, errors.Is(err, cases[i].err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
