package tree_test

import (
	"encoding/json"
	"fmt"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxTree(t *testing.T) {
	zeroTransfer := new(intMaxTypes.Transfer).SetZero()
	transferTree, err := intMaxTree.NewTransferTree(7, nil, zeroTransfer.Hash())
	assert.NoError(t, err)
	transferTreeRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()
	zeroTx, err := intMaxTypes.NewTx(
		&transferTreeRoot,
		0,
	)
	require.Nil(t, err)
	zeroTxHash := zeroTx.Hash()
	initialLeaves := make([]*intMaxTypes.Tx, 0)
	mt, err := intMaxTree.NewTxTree(3, initialLeaves, zeroTxHash)
	require.Nil(t, err)

	leaves := make([]*intMaxTypes.Tx, 8)
	for i := 0; i < 4; i++ {
		leaves[i] = new(intMaxTypes.Tx).Set(zeroTx)
		leaves[i].Nonce = uint64(i)
		require.Nil(t, err)
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	expectedRoot := intMaxGP.Compress(
		intMaxGP.Compress(intMaxGP.Compress(leaves[0].Hash(), leaves[1].Hash()), intMaxGP.Compress(leaves[2].Hash(), leaves[3].Hash())),
		intMaxGP.Compress(intMaxGP.Compress(zeroTxHash, zeroTxHash), intMaxGP.Compress(zeroTxHash, zeroTxHash)),
	)
	// expectedRoot :=
	// 	intMaxGP.Compress(intMaxGP.Compress(leaves[0].Hash(), leaves[1].Hash()), intMaxGP.Compress(leaves[2].Hash(), leaves[3].Hash()))
	actualRoot, _, _ := mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)

	leaves[4] = new(intMaxTypes.Tx).Set(zeroTx)
	leaves[4].Nonce = uint64(4)
	assert.Nil(t, err)
	_, err = mt.AddLeaf(4, leaves[4])
	require.Nil(t, err)

	expectedRoot = intMaxGP.Compress(
		intMaxGP.Compress(intMaxGP.Compress(leaves[0].Hash(), leaves[1].Hash()), intMaxGP.Compress(leaves[2].Hash(), leaves[3].Hash())),
		intMaxGP.Compress(intMaxGP.Compress(leaves[4].Hash(), zeroTxHash), intMaxGP.Compress(zeroTxHash, zeroTxHash)),
	)
	actualRoot, _, _ = mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)

	for i := 5; i < 8; i++ {
		leaves[i] = new(intMaxTypes.Tx).Set(zeroTx)
		leaves[i].Nonce = uint64(i)
		assert.Nil(t, err)
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	expectedRoot = intMaxGP.Compress(
		intMaxGP.Compress(intMaxGP.Compress(leaves[0].Hash(), leaves[1].Hash()), intMaxGP.Compress(leaves[2].Hash(), leaves[3].Hash())),
		intMaxGP.Compress(intMaxGP.Compress(leaves[4].Hash(), leaves[5].Hash()), intMaxGP.Compress(leaves[6].Hash(), leaves[7].Hash())),
	)
	actualRoot, _, _ = mt.GetCurrentRootCountAndSiblings()
	assert.Equal(t, expectedRoot.Elements, actualRoot.Elements)
}

func TestWithdrawalRequest(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	transfers := make([]*intMaxTypes.Transfer, 8)

	for i := 0; i < 8; i++ {
		addressHex := "0x25817fFA38D884A93F22154EEE61E6D1533B73E2"
		address, err := hexutil.Decode(addressHex)
		assert.NoError(t, err)
		recipient, err := intMaxTypes.NewEthereumAddress(address)
		assert.NoError(t, err)
		assert.NotNil(t, recipient)
		amount := new(big.Int).Rand(r, big.NewInt(100000))
		assert.NoError(t, err)
		salt := new(intMaxTree.PoseidonHashOut)
		salt.Elements[0] = *new(ffg.Element).SetUint64(1)
		salt.Elements[1] = *new(ffg.Element).SetUint64(2)
		salt.Elements[2] = *new(ffg.Element).SetUint64(3)
		salt.Elements[3] = *new(ffg.Element).SetUint64(4)
		transferData := intMaxTypes.Transfer{
			Recipient:  recipient,
			TokenIndex: 0,
			Amount:     amount,
			Salt:       salt,
		}
		transfers[i] = &transferData
	}

	zeroHash := intMaxGP.NewPoseidonHashOut()
	transferTree, err := intMaxTree.NewTransferTree(6, transfers, zeroHash)
	assert.NoError(t, err)

	var transferIndex uint64 = 2
	transferMerkleProof, transferTreeRoot, err := transferTree.ComputeMerkleProof(transferIndex)
	require.Nil(t, err)
	// expectedRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()
	// assert.Equal(t, expectedRoot.Elements, transferTreeRoot.Elements)

	fmt.Printf("%v\n", transferMerkleProof)

	zeroTx, err := intMaxTypes.NewTx(
		zeroHash,
		0,
	)
	require.Nil(t, err)

	zeroTxHash := zeroTx.Hash()
	initialLeaves := make([]*intMaxTypes.Tx, 0)
	mt, err := intMaxTree.NewTxTree(7, initialLeaves, zeroTxHash)
	require.Nil(t, err)

	leaves := make([]*intMaxTypes.Tx, 8)
	for i := 0; i < 4; i++ {
		leaves[i] = new(intMaxTypes.Tx).Set(zeroTx)
		leaves[i].Nonce = uint64(i)
		require.Nil(t, err)
		_, err := mt.AddLeaf(uint64(i), leaves[i])
		require.Nil(t, err)
	}

	leaves[4] = new(intMaxTypes.Tx).Set(zeroTx)
	leaves[4].Nonce = uint64(4)
	_, err = mt.AddLeaf(4, leaves[4])
	require.Nil(t, err)

	leaves[5] = new(intMaxTypes.Tx).Set(zeroTx)
	leaves[5].Nonce = uint64(5)
	_, err = mt.AddLeaf(5, leaves[5])
	require.Nil(t, err)

	var transactionIndex uint64 = 4
	txMerkleProof, txTreeRoot, err := mt.ComputeMerkleProof(transactionIndex)
	expectedRoot, _, _ := mt.GetCurrentRootCountAndSiblings()
	require.Nil(t, err)
	assert.Equal(t, expectedRoot.Elements, txTreeRoot.Elements)

	fmt.Printf("txMerkleProof: %v\n", txMerkleProof)

	tx4 := intMaxTypes.Tx{
		TransferTreeRoot: &transferTreeRoot,
		Nonce:            leaves[4].Nonce,
	}

	enoughBalanceProof := intMaxTypes.Plonky2Proof{
		Proof:        []byte{},
		PublicInputs: []ffg.Element{},
	}

	withdrawalRequest := WithdrawalRequest{
		TransferData: *transfers[transferIndex],
		TransferMerkleProof: MerkleProofWithIndex{
			Siblings: transferMerkleProof,
			Index:    transferIndex,
		},
		Transaction: tx4,
		TxMerkleProof: MerkleProofWithIndex{
			Siblings: txMerkleProof,
			Index:    transactionIndex,
		},
		TxTreeRoot:         txTreeRoot,
		EnoughBalanceProof: enoughBalanceProof,
	}

	withdrawalRequestJson, err := json.Marshal(withdrawalRequest)
	require.Nil(t, err)
	fmt.Printf("withdrawalRequestJson: %v\n", string(withdrawalRequestJson))
}

type MerkleProofWithIndex struct {
	Siblings []*intMaxGP.PoseidonHashOut `json:"siblings"`
	Index    uint64                      `json:"index"`
}

type WithdrawalRequest struct {
	TransferData        intMaxTypes.Transfer     `json:"transferData"`
	TransferMerkleProof MerkleProofWithIndex     `json:"transferMerkleProof"`
	Transaction         intMaxTypes.Tx           `json:"transaction"`
	TxMerkleProof       MerkleProofWithIndex     `json:"txMerkleProof"`
	TxTreeRoot          intMaxGP.PoseidonHashOut `json:"txTreeRoot"`
	EnoughBalanceProof  intMaxTypes.Plonky2Proof `json:"enoughBalanceProof"`
}
