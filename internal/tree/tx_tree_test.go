package tree

import (
	"intmax2-node/internal/hash/goldenposeidon"
	"testing"

	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestTxTree(t *testing.T) {
	zeroTx := Tx{
		FeeTransferHash:  goldenposeidon.NewPoseidonHashOut(),
		TransferTreeRoot: goldenposeidon.NewPoseidonHashOut(),
	}
	zeroTxHash := zeroTx.Hash()
	initialLeaves := make([]*Tx, 0)
	mt, err := NewTxTree(3, initialLeaves, zeroTxHash)
	require.Nil(t, err)

	leaves := make([]*Tx, 8)
	for i := 0; i < 4; i++ {
		leaves[i], _ = new(Tx).SetRandom()
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	expectedRoot := goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0].Hash(), leaves[1].Hash()), goldenposeidon.Compress(leaves[2].Hash(), leaves[3].Hash())),
		goldenposeidon.Compress(goldenposeidon.Compress(zeroTxHash, zeroTxHash), goldenposeidon.Compress(zeroTxHash, zeroTxHash)),
	)
	// expectedRoot :=
	// 	goldenposeidon.Compress(goldenposeidon.Compress(leaves[0].Hash(), leaves[1].Hash()), goldenposeidon.Compress(leaves[2].Hash(), leaves[3].Hash()))
	actualRoot, _, _ := mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)

	leaves[4], _ = new(Tx).SetRandom()
	_, err = mt.AddLeaf(4, leaves[4])
	require.Nil(t, err)

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0].Hash(), leaves[1].Hash()), goldenposeidon.Compress(leaves[2].Hash(), leaves[3].Hash())),
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[4].Hash(), zeroTxHash), goldenposeidon.Compress(zeroTxHash, zeroTxHash)),
	)
	actualRoot, _, _ = mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)

	for i := 5; i < 8; i++ {
		leaves[i], _ = new(Tx).SetRandom()
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	expectedRoot = goldenposeidon.Compress(
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[0].Hash(), leaves[1].Hash()), goldenposeidon.Compress(leaves[2].Hash(), leaves[3].Hash())),
		goldenposeidon.Compress(goldenposeidon.Compress(leaves[4].Hash(), leaves[5].Hash()), goldenposeidon.Compress(leaves[6].Hash(), leaves[7].Hash())),
	)
	actualRoot, _, _ = mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)
}
