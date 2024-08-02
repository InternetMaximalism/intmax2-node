package types_test

import (
	"crypto/rand"
	"encoding/base64"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
	"intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// {
// 	"nonce": 2,
// 	"powNonce": "0xc1",
// 	"transferData": [
// 	  {
// 		"amount": "10",
// 		"salt": "0x0000000000000000000000000000000000000000000000000000000000000001",
// 		"tokenIndex": "0",
// 		"recipient": {
// 		  "address_type": "INTMAX",
// 		  "address": "0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3"
// 		}
// 	  },
// 	  {
// 		"amount": "10",
// 		"salt": "0x0000000000000000000000000000000000000000000000000000000000000002",
// 		"tokenIndex": "0",
// 		"recipient": {
// 		  "address_type": "ETHEREUM",
// 		  "address": "0xD7fa191fB4F255f7Af801966819382edDA19E09C"
// 		}
// 	  }
// 	]
// }

func TestTxHash(t *testing.T) {
	blockBuilderAddress, err := hexutil.Decode("0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3")
	assert.NoError(t, err)
	blockBuilderGenericAddress, err := intMaxTypes.NewINTMAXAddress(blockBuilderAddress)
	assert.NoError(t, err)
	salt := goldenposeidon.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")

	recipient, err := hexutil.Decode("0xD7fa191fB4F255f7Af801966819382edDA19E09C")
	assert.NoError(t, err)
	recipientGenericAddress, err := intMaxTypes.NewEthereumAddress(recipient)
	assert.NoError(t, err)

	const numTxTreeLeaves = 128
	transfers := [numTxTreeLeaves]intMaxTypes.Transfer{}
	transfers[0] = intMaxTypes.Transfer{
		Amount:     big.NewInt(10),
		Salt:       salt,
		TokenIndex: 0,
		Recipient:  blockBuilderGenericAddress,
	}
	transfers[1] = intMaxTypes.Transfer{
		Amount:     big.NewInt(10),
		Salt:       goldenposeidon.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
		TokenIndex: 0,
		Recipient:  recipientGenericAddress,
	}

	transferTree, err := tree.NewTransferTree(7, nil, new(intMaxTypes.Transfer).SetZero().Hash())
	assert.NoError(t, err)
	transferTreeRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	tx, err := intMaxTypes.NewTx(
		&transferTreeRoot,
		1,
	)
	assert.NoError(t, err)

	txHash := tx.Hash()
	assert.Equal(t, "0x378999af8ce0013df99b58c799161f711150fa56c8255c432235a2e0b9fd605f", txHash.String())
}

func TestEncryptTxDetails(t *testing.T) {
	senderAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	require.NoError(t, err)
	recipientAccount1, err := intMaxAcc.NewPrivateKey(big.NewInt(4))
	require.NoError(t, err)
	recipient1, err := intMaxTypes.NewINTMAXAddress(recipientAccount1.ToAddress().Bytes())
	require.NoError(t, err)
	recipientAccount2, err := intMaxAcc.NewPrivateKey(big.NewInt(5))
	require.NoError(t, err)
	recipient2, err := intMaxTypes.NewINTMAXAddress(recipientAccount2.ToAddress().Bytes())
	require.NoError(t, err)
	recipientAccount3, err := intMaxAcc.NewPrivateKey(big.NewInt(6))
	require.NoError(t, err)
	recipient3, err := intMaxTypes.NewINTMAXAddress(recipientAccount3.ToAddress().Bytes())
	require.NoError(t, err)

	salt := new(goldenposeidon.PoseidonHashOut)
	saltBytes := make([]byte, 32)
	_, err = rand.Read(saltBytes)
	require.NoError(t, err)

	transfers := []*intMaxTypes.Transfer{
		{
			Recipient:  recipient1,
			TokenIndex: 0,
			Amount:     big.NewInt(100),
			Salt:       salt,
		},
		{
			Recipient:  recipient2,
			TokenIndex: 1,
			Amount:     big.NewInt(200),
			Salt:       salt,
		},
		{
			Recipient:  recipient3,
			TokenIndex: 0,
			Amount:     big.NewInt(300),
			Salt:       salt,
		},
	}

	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	transferTree, err := tree.NewTransferTree(7, transfers, zeroTransfer.Hash())
	require.NoError(t, err)

	TransferTreeRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	tx := intMaxTypes.Tx{
		Nonce:            2,
		TransferTreeRoot: &TransferTreeRoot,
	}

	txDetails := intMaxTypes.TxDetails{
		Tx:        tx,
		Transfers: transfers,
	}

	encodedTx := txDetails.Marshal()

	encryptedTransfer, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		senderAccount.Public(),
		encodedTx,
	)
	require.NoError(t, err)

	encodedText := base64.StdEncoding.EncodeToString(encryptedTransfer)

	t.Log("encodedTransfer", encodedText)

	decodedText, err := base64.StdEncoding.DecodeString(encodedText)
	require.NoError(t, err)

	decryptedTxBytes, err := senderAccount.DecryptECIES(
		decodedText,
	)
	require.NoError(t, err)
	require.Equal(t, encodedTx, decryptedTxBytes)

	decryptedTx := new(intMaxTypes.TxDetails)

	err = decryptedTx.Unmarshal(decryptedTxBytes)
	require.NoError(t, err)
	assert.True(
		t, txDetails.Tx.Equal(&decryptedTx.Tx),
		"recipients should be equal: %+v != %+v", txDetails, decryptedTx,
	)
}
