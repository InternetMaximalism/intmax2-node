package types_test

import (
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
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
	tx := intMaxTypes.Tx{
		Nonce:     1,
		PowNonce:  0xc1,
		Transfers: transfers,
	}

	txHash := tx.Hash()
	assert.Equal(t, "0x6c924e688516950a28f33af9435ddb279a84e7934421eed491d29cdf73ec5e5d", txHash.String())
}

// func TestRandomTxHash(t *testing.T) {
// 	// transfersHash, err := new(intMaxTypes.PoseidonHashOut).SetRandom()
// 	// assert.Nil(t, err)
// 	initialLeaves := make([]*intMaxTypes.Transfer, 2)

// 	tx := intMaxTypes.Tx{
// 		Nonce:     1,
// 		PowNonce:  0xc1,
// 		Transfers: []intMaxTypes.Transfer{},
// 	}
// 	zeroHash := intMaxTypes.Tx{
// 		Nonce:     0,
// 		PowNonce:  0,
// 		Transfers: []intMaxTypes.Transfer{},
// 	}
// 	var height uint8 = 7
// 	txTreeRoot, err := tree.NewTransferTree(height, initialLeaves, zeroHash)
// 	assert.Nil(t, err)

// 	permutedTx := goldenposeidon.Permute([12]*ffg.Element{
// 		new(ffg.Element).Set(&txTreeRoot.Elements[0]),
// 		new(ffg.Element).Set(&txTreeRoot.Elements[1]),
// 		new(ffg.Element).Set(&txTreeRoot.Elements[2]),
// 		new(ffg.Element).Set(&txTreeRoot.Elements[3]),
// 		new(ffg.Element).SetUint64(tx.Nonce),
// 		new(ffg.Element).SetUint64(tx.PowNonce),
// 		new(ffg.Element).SetZero(),
// 		new(ffg.Element).SetZero(),
// 		new(ffg.Element).SetZero(),
// 		new(ffg.Element).SetZero(),
// 		new(ffg.Element).SetZero(),
// 		new(ffg.Element).SetZero(),
// 	})
// 	expected := intMaxTypes.PoseidonHashOut{
// 		Elements: [4]ffg.Element{
// 			*permutedTx[0],
// 			*permutedTx[1],
// 			*permutedTx[2],
// 			*permutedTx[3],
// 		},
// 	}

// 	txHash := tx.Hash()
// 	assert.Equal(t, expected.String(), txHash.String())
// }
