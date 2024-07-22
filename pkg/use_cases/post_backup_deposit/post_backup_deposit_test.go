package post_backup_deposit_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/mnemonic_wallet"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/backup_deposit"
	ucPostBackupDeposit "intmax2-node/pkg/use_cases/post_backup_deposit"
	"intmax2-node/pkg/use_cases/post_backup_transfer"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUseCasePostBackupTransactionTest(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := ucPostBackupDeposit.New()

	const (
		senderKey      = "sender"
		successMessage = "Deposit data backup successful."
	)

	sampleInput := MakeSamplePostBackupDepositRequest(t, configs.New())

	cases := []struct {
		desc  string
		input *backup_deposit.UCPostBackupDepositInput
		err   error
	}{
		{
			desc: fmt.Sprintf("Error: %s", ucPostBackupDeposit.ErrUCPostBackupDepositInputEmpty),
			err:  ucPostBackupDeposit.ErrUCPostBackupDepositInputEmpty,
		},
		{
			desc:  "Success",
			input: sampleInput,
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
				assert.Equal(t, resp.Message, successMessage)
			}
		})
	}
}

func MakeSamplePostBackupDepositRequest(t *testing.T, cfg *configs.Config) *backup_deposit.UCPostBackupDepositInput {
	const (
		mnemonic         = "gown situate miss skill figure rain smoke grief giraffe perfect milk gospel casino open mimic egg grace canoe erode skull drip open luggage next"
		mnPassword       = ""
		derivation       = "m/44'/60'/0'/0/0"
		senderAddressHex = "0x1c6f2045ddc7fde4f0ff37ac47b2726ed2e6e9fe8ea3d3d6971403cece12306d"
	)

	w, err := mnemonic_wallet.New().WalletFromMnemonic(mnemonic, mnPassword, derivation)
	assert.NoError(t, err)
	assert.Equal(t, w.IntMaxWalletAddress, senderAddressHex)

	recipientAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(4))
	assert.NoError(t, err)
	recipientAddress, err := intMaxTypes.NewINTMAXAddress(recipientAccount.ToAddress().Bytes())
	assert.NoError(t, err)

	salt := new(intMaxTypes.PoseidonHashOut).SetZero()
	deposit := intMaxTypes.Transfer{
		Recipient:  recipientAddress,
		TokenIndex: 1,
		Amount:     big.NewInt(100),
		Salt:       salt,
	}

	plaintext := bytes.NewBuffer(make([]byte, 0))
	err = post_backup_transfer.WriteTransfer(plaintext, &deposit)
	assert.NoError(t, err)
	encryptedDeposit, err := intMaxAcc.EncryptECIES(rand.Reader, recipientAccount.Public(), plaintext.Bytes())
	assert.NoError(t, err)

	const blockNumber uint32 = 1

	return &backup_deposit.UCPostBackupDepositInput{
		Recipient:        recipientAccount.ToAddress().String(),
		DecodeRecipient:  recipientAccount.Public(),
		BlockNumber:      blockNumber,
		EncryptedDeposit: hexutil.Encode(encryptedDeposit), // TODO: Base64
	}
}
