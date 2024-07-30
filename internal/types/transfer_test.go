package types_test

import (
	"crypto/rand"
	"encoding/base64"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"testing"

	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransferData(t *testing.T) {
	address := make([]byte, 32)
	for i := 0; i < 32; i++ {
		address[i] = byte(i)
	}
	recipient, err := intMaxTypes.NewINTMAXAddress(address)
	assert.NoError(t, err)
	assert.NotNil(t, recipient)
	amount := new(big.Int).SetUint64(100)
	assert.NoError(t, err)
	salt := new(intMaxTypes.PoseidonHashOut)
	salt.Elements[0] = *new(ffg.Element).SetUint64(1)
	salt.Elements[1] = *new(ffg.Element).SetUint64(2)
	salt.Elements[2] = *new(ffg.Element).SetUint64(3)
	salt.Elements[3] = *new(ffg.Element).SetUint64(4)
	transferData := intMaxTypes.Transfer{
		Recipient:  recipient,
		TokenIndex: 1,
		Amount:     amount,
		Salt:       salt,
	}

	flattenedTransfer := transferData.Marshal()
	assert.Equal(t, 100, len(flattenedTransfer))

	transferHash := transferData.Hash()
	assert.Equal(t, transferHash.String(), "0x448cb4641023c5c62338428ee32e0f267d33246cc7275e1b9f994236d92870d7")
}

func TestEncryptTransfers(t *testing.T) {
	senderAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	require.NoError(t, err)
	recipientAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(4))
	require.NoError(t, err)
	recipient, err := intMaxTypes.NewINTMAXAddress(recipientAccount.ToAddress().Bytes())
	require.NoError(t, err)

	salt := new(goldenposeidon.PoseidonHashOut)
	saltBytes := make([]byte, 32)
	_, err = rand.Read(saltBytes)
	require.NoError(t, err)

	transfer := intMaxTypes.Transfer{
		Recipient:  recipient,
		TokenIndex: 0,
		Amount:     big.NewInt(100),
		Salt:       salt,
	}

	encodedTransfer := transfer.Marshal()
	// t.Log("encodedTransfer", len(encodedTransfer))

	encryptedTransfer, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		senderAccount.Public(),
		encodedTransfer,
	)
	require.NoError(t, err)

	encodedText := base64.StdEncoding.EncodeToString(encryptedTransfer)

	// t.Log("encodedTransfer", encodedText)

	decodedText, err := base64.StdEncoding.DecodeString(encodedText)
	require.NoError(t, err)

	decryptedTransferBytes, err := senderAccount.DecryptECIES(
		decodedText,
	)
	require.NoError(t, err)
	// t.Log("decryptedDepositBytes", len(decryptedTransferBytes))
	require.Equal(t, encodedTransfer, decryptedTransferBytes)

	decryptedDeposit := new(intMaxTypes.Transfer)

	err = decryptedDeposit.Unmarshal(decryptedTransferBytes)
	require.NoError(t, err)
	assert.True(
		t, transfer.Equal(decryptedDeposit),
		"recipients should be equal: %v != %v", transfer, decryptedDeposit,
	)
}
