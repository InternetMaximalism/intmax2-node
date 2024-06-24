package types

import (
	"intmax2-node/internal/hash/goldenposeidon"
	"testing"

	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
)

func TestTxHash(t *testing.T) {
	tx := Tx{
		FeeTransferHash:  goldenposeidon.HexToHash("0x83fc198de31e1b2b1a8212d2430fbb7766c13d9ad305637dea3759065606475d"),
		TransferTreeRoot: goldenposeidon.HexToHash("0xb32f96fad8af99f3b3cb90dfbb4849f73435dbee1877e4ac2c213127379549ce"),
	}

	txHash := tx.Hash()
	assert.Equal(t, "0xbe806cb0804af0523bc85ba5cbf03d12dc6579fe5c2d6670b6421571473e0e4d", txHash.String())
}

func TestRandomTxHash(t *testing.T) {
	feeTransferHash, err := new(poseidonHashOut).SetRandom()
	assert.Nil(t, err)
	transferTreeRoot, err := new(poseidonHashOut).SetRandom()
	assert.Nil(t, err)
	tx := Tx{
		FeeTransferHash:  feeTransferHash,
		TransferTreeRoot: transferTreeRoot,
	}
	txHash := tx.Hash()

	permutedTx := goldenposeidon.Permute([12]*ffg.Element{
		new(ffg.Element).Set(&tx.FeeTransferHash.Elements[0]),
		new(ffg.Element).Set(&tx.FeeTransferHash.Elements[1]),
		new(ffg.Element).Set(&tx.FeeTransferHash.Elements[2]),
		new(ffg.Element).Set(&tx.FeeTransferHash.Elements[3]),
		new(ffg.Element).Set(&tx.TransferTreeRoot.Elements[0]),
		new(ffg.Element).Set(&tx.TransferTreeRoot.Elements[1]),
		new(ffg.Element).Set(&tx.TransferTreeRoot.Elements[2]),
		new(ffg.Element).Set(&tx.TransferTreeRoot.Elements[3]),
		new(ffg.Element).SetZero(),
		new(ffg.Element).SetZero(),
		new(ffg.Element).SetZero(),
		new(ffg.Element).SetZero(),
	})
	expected := poseidonHashOut{
		Elements: [4]ffg.Element{
			*permutedTx[0],
			*permutedTx[1],
			*permutedTx[2],
			*permutedTx[3],
		},
	}
	assert.Equal(t, expected.String(), txHash.String())
}
