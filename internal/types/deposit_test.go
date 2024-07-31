package types_test

import (
	"crypto/rand"
	"encoding/base64"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDeposit(t *testing.T) {
	recipientAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(4))
	require.NoError(t, err)

	saltBytes := make([]byte, 32)
	_, err = rand.Read(saltBytes)
	require.NoError(t, err)

	salt := new(goldenposeidon.PoseidonHashOut)
	err = salt.Unmarshal(saltBytes)
	require.NoError(t, err)

	deposit := intMaxTypes.Deposit{
		Recipient:  recipientAccount.Public(),
		TokenIndex: 0,
		Amount:     big.NewInt(300),
		Salt:       salt,
	}

	encodedDeposit := deposit.Marshal()
	encryptedDeposit, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		recipientAccount.Public(),
		encodedDeposit,
	)
	require.NoError(t, err)

	encodedText := base64.StdEncoding.EncodeToString(encryptedDeposit)

	t.Log("encodedDeposit", encodedText)

	decodedText, err := base64.StdEncoding.DecodeString(encodedText)
	require.NoError(t, err)

	decryptedDepositBytes, err := recipientAccount.DecryptECIES(
		decodedText,
	)
	require.NoError(t, err)

	decryptedDeposit := new(intMaxTypes.Deposit)

	err = decryptedDeposit.Unmarshal(decryptedDepositBytes)
	assert.NoError(t, err)
	assert.True(
		t, deposit.Equal(decryptedDeposit),
		"recipients should be equal: %v != %v", deposit, decryptedDeposit,
	)
}
