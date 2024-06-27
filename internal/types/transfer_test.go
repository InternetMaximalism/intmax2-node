package types_test

import (
	"fmt"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"testing"

	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
)

func TestTransferData(t *testing.T) {
	address := make([]byte, 32)
	for i := 0; i < 32; i++ {
		address[i] = byte(i)
	}
	recipient, err := intMaxTypes.NewINTMAXAddress(address)
	assert.NoError(t, err)
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
	fmt.Printf("transferData: %v\n", transferData)

	flattenedTransfer := transferData.Marshal()
	fmt.Printf("flattenedTransfer: %v\n", flattenedTransfer)
	assert.Equal(t, 100, len(flattenedTransfer))

	transferHash := transferData.Hash()
	assert.Equal(t, transferHash.String(), "0xe01a5851b48f1e3affcc03823946c4b6e843caa29ba8d9ee77d3617048c683ac")
}
