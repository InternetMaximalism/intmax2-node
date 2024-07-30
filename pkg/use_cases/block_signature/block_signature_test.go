package block_signature_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	blockSignature "intmax2-node/internal/use_cases/block_signature"
	ucBlockSignature "intmax2-node/pkg/use_cases/block_signature"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var ErrCannotOpenBinaryFile = errors.New("cannot open binary file")

var ErrCannotGetFileInformation = errors.New("cannot get file information")

var ErrCannotReadBinaryFile = errors.New("cannot read binary file")

var ErrCannotParseJson = errors.New("cannot parse JSON")

func TestUseCaseTransaction(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := NewMockWorker(ctrl)

	cfg := new(configs.Config)
	uc := ucBlockSignature.New(cfg, w)

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

	publicInputs, err := readPlonky2PublicInputsJson("balance_proof_public_inputs.json")
	assert.NoError(t, err)
	proof, err := readPlonky2ProofBinary("balance_proof.bin")
	assert.NoError(t, err)

	enoughBalanceProof := &blockSignature.EnoughBalanceProofInput{
		PrevBalanceProof: &blockSignature.Plonky2Proof{
			PublicInputs: publicInputs,
			Proof:        proof,
		},
		TransferStepProof: &blockSignature.Plonky2Proof{
			PublicInputs: []uint64{
				2726224824249046055, 14025881618846813748, 5361314524880173070, 2912484915938769214,
			},
			Proof: []byte{0x99, 0x39, 0x6b, 0x28},
		}, // dummy
	}
	wrongEnoughBalanceProof := new(blockSignature.EnoughBalanceProofInput).Set(enoughBalanceProof)
	wrongEnoughBalanceProof.PrevBalanceProof.PublicInputs[0] = 2726224824249046055

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
			err = uc.Do(ctx, cases[i].input)
			if cases[i].err != nil {
				assert.True(t, errors.Is(err, cases[i].err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func readPlonky2ProofBinary(filePath string) ([]byte, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Join(ErrCannotOpenBinaryFile, err)
	}
	defer func() {
		_ = file.Close()
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Join(ErrCannotGetFileInformation, err)
	}

	// Create buffer
	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)

	// Read file content
	_, err = file.Read(buffer)
	if err != nil {
		return nil, errors.Join(ErrCannotReadBinaryFile, err)
	}

	return buffer, nil
}

func readPlonky2PublicInputsJson(filePath string) ([]uint64, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Join(ErrCannotOpenBinaryFile, err)
	}
	defer func() {
		_ = file.Close()
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Join(ErrCannotGetFileInformation, err)
	}

	// Create buffer
	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)

	// Read file content
	_, err = file.Read(buffer)
	if err != nil {
		return nil, errors.Join(ErrCannotReadBinaryFile, err)
	}

	// Parse JSON
	var publicInputsStr []string
	err = json.Unmarshal(buffer, &publicInputsStr)
	if err != nil {
		return nil, errors.Join(ErrCannotParseJson, err)
	}

	publicInputs := make([]uint64, len(publicInputsStr))
	for i, publicInputStr := range publicInputsStr {
		publicInput, err := strconv.ParseUint(publicInputStr, 10, 64)
		if err != nil {
			return nil, errors.Join(ErrCannotParseJson, err)
		}
		publicInputs[i] = publicInput
	}

	return publicInputs, nil
}
