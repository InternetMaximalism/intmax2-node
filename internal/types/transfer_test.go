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
	a, _ := new(big.Int).SetString("5072999951032826783367862081641321578167449493857840371024146051846401100402", 10)
	copy(address[32-len(a.Bytes()):], a.Bytes())
	recipient, err := intMaxTypes.NewINTMAXAddress(address)
	assert.NoError(t, err)
	assert.NotNil(t, recipient)
	amount, _ := new(big.Int).SetString("4098227373595779913487602621708067623280365761755196974852", 10)
	assert.NoError(t, err)
	salt := new(intMaxTypes.PoseidonHashOut)
	salt.Elements[0] = *new(ffg.Element).SetUint64(3645147740416515513)
	salt.Elements[1] = *new(ffg.Element).SetUint64(16630279128197546175)
	salt.Elements[2] = *new(ffg.Element).SetUint64(5615096774476642530)
	salt.Elements[3] = *new(ffg.Element).SetUint64(7135915778368184935)
	transferData := intMaxTypes.Transfer{
		Recipient:  recipient,
		TokenIndex: 827790650,
		Amount:     amount,
		Salt:       salt,
	}

	flattenedTransfer := transferData.Marshal()
	assert.Equal(t, 100, len(flattenedTransfer))

	t.Log("transferData.ToUint64Slice()", transferData.ToUint64Slice())

	transferHash := transferData.Hash()
	assert.Equal(t, transferHash.String(), "0x9b469cf21563c28179698ccac6a789450e68b270b586e6ec583cead158f44631")
}

func TestTransferDetails(t *testing.T) {
	address := make([]byte, 32)
	for i := 0; i < 32; i++ {
		address[i] = byte(i)
	}
	a, _ := new(big.Int).SetString("5072999951032826783367862081641321578167449493857840371024146051846401100402", 10)
	copy(address[32-len(a.Bytes()):], a.Bytes())
	recipient, err := intMaxTypes.NewINTMAXAddress(address)
	assert.NoError(t, err)
	assert.NotNil(t, recipient)
	amount, _ := new(big.Int).SetString("4098227373595779913487602621708067623280365761755196974852", 10)
	assert.NoError(t, err)
	salt := new(intMaxTypes.PoseidonHashOut)
	salt.Elements[0] = *new(ffg.Element).SetUint64(3645147740416515513)
	salt.Elements[1] = *new(ffg.Element).SetUint64(16630279128197546175)
	salt.Elements[2] = *new(ffg.Element).SetUint64(5615096774476642530)
	salt.Elements[3] = *new(ffg.Element).SetUint64(7135915778368184935)
	transferData := intMaxTypes.Transfer{
		Recipient:  recipient,
		TokenIndex: 827790650,
		Amount:     amount,
		Salt:       salt,
	}

	transferDetails := intMaxTypes.TransferDetails{
		TransferWitness: &intMaxTypes.TransferWitness{
			Transfer:            transferData,
			TransferIndex:       0,
			TransferMerkleProof: make([]*goldenposeidon.PoseidonHashOut, 6),
			Tx: intMaxTypes.Tx{
				Nonce:            0,
				TransferTreeRoot: new(goldenposeidon.PoseidonHashOut).SetZero(),
			},
		},
		TxTreeRoot:                new(goldenposeidon.PoseidonHashOut).SetZero(),
		TxMerkleProof:             make([]*goldenposeidon.PoseidonHashOut, 7),
		SenderBalancePublicInputs: []byte{},
	}
	for i := 0; i < 6; i++ {
		transferDetails.TransferWitness.TransferMerkleProof[i] = new(goldenposeidon.PoseidonHashOut).SetZero()
	}
	for i := 0; i < 7; i++ {
		transferDetails.TxMerkleProof[i] = new(goldenposeidon.PoseidonHashOut).SetZero()
	}

	encodedTransferDetails := transferDetails.Marshal()

	t.Log("transferDetails.Marshal()", encodedTransferDetails)
	t.Log("transferDetails.Marshal()", len(encodedTransferDetails))

	decodedTransferDetails := new(intMaxTypes.TransferDetails)
	err = decodedTransferDetails.Unmarshal(encodedTransferDetails)
	require.NoError(t, err)

	require.True(t, transferDetails.Equal(decodedTransferDetails))
}

func TestEncryptTransfers(t *testing.T) {
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
		Amount:     big.NewInt(500),
		Salt:       salt,
	}

	encodedTransfer := transfer.Marshal()

	encryptedTransfer, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		recipientAccount.Public(),
		encodedTransfer,
	)
	require.NoError(t, err)

	encodedText := base64.StdEncoding.EncodeToString(encryptedTransfer)

	t.Log("encodedTransfer", encodedText)

	decodedText, err := base64.StdEncoding.DecodeString(encodedText)
	require.NoError(t, err)

	decryptedTransferBytes, err := recipientAccount.DecryptECIES(
		decodedText,
	)
	require.NoError(t, err)
	require.Equal(t, encodedTransfer, decryptedTransferBytes)

	decryptedTransfer := new(intMaxTypes.Transfer)

	err = decryptedTransfer.Unmarshal(decryptedTransferBytes)
	require.NoError(t, err)
	assert.True(
		t, transfer.Equal(decryptedTransfer),
		"transfer should be equal: %v != %v", transfer, decryptedTransfer,
	)
}
