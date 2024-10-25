package block_validity_prover

import (
	"intmax2-node/internal/accounts"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type MockTxRequest struct {
	Sender              *intMaxAcc.PrivateKey
	AccountID           uint64
	Tx                  *intMaxTypes.Tx
	WillReturnSignature bool
}

// TODO: refactor this function
func NewBlockContentFromTxRequests(isRegistrationBlock bool, txs []*MockTxRequest) (*intMaxTypes.BlockContent, error) {
	const numOfSenders = 128
	if len(txs) > numOfSenders {
		panic("too many txs")
	}

	// sort and pad txs
	sortedTxs := make([]*MockTxRequest, len(txs))
	copy(sortedTxs, txs)
	sort.Slice(sortedTxs, func(i, j int) bool {
		return sortedTxs[j].Sender.PublicKey.BigInt().Cmp(sortedTxs[i].Sender.PublicKey.BigInt()) == 1
	})

	publicKeys := make([]*accounts.PublicKey, len(sortedTxs))
	for i, tx := range sortedTxs {
		publicKeys[i] = tx.Sender.Public()
	}

	dummyPublicKey := accounts.NewDummyPublicKey()
	for i := len(publicKeys); i < numOfSenders; i++ {
		publicKeys = append(publicKeys, dummyPublicKey)
	}

	zeroTx := new(intMaxTypes.Tx).SetZero()
	txTree, err := tree.NewTxTree(uint8(tree.TX_TREE_HEIGHT), nil, zeroTx.Hash())
	if err != nil {
		panic(err)
	}

	for _, tx := range txs {
		_, index, _ := txTree.GetCurrentRootCountAndSiblings()
		_, err = txTree.AddLeaf(index, tx.Tx)
		if err != nil {
			panic(err)
		}
	}

	txTreeRoot, _, _ := txTree.GetCurrentRootCountAndSiblings()

	flattenTxTreeRoot := finite_field.BytesToFieldElementSlice(txTreeRoot.Marshal())

	addresses := make([]intMaxTypes.Uint256, len(publicKeys))
	for i, publicKey := range publicKeys {
		addresses[i] = *new(intMaxTypes.Uint256).FromBigInt(publicKey.BigInt())
	}
	publicKeysHash := GetPublicKeysHash(addresses)

	signatures := make([]*bn254.G2Affine, len(sortedTxs))
	for i, keyPair := range sortedTxs {
		var signature *bn254.G2Affine
		signature, err = keyPair.Sender.WeightByHash(publicKeysHash.Bytes()).Sign(flattenTxTreeRoot)
		if err != nil {
			return nil, err
		}
		signatures[i] = signature
	}

	encodedSignatures := make([]string, len(sortedTxs))
	for i, signature := range signatures {
		encodedSignatures[i] = hexutil.Encode(signature.Marshal())
	}

	var blockContent *intMaxTypes.BlockContent
	blockContent, err = block_post_service.MakeRegistrationBlock(txTreeRoot, publicKeys, encodedSignatures)
	if err != nil {
		return nil, err
	}

	return blockContent, nil
}

func GetPublicKeysHash(publicKeys []intMaxTypes.Uint256) intMaxTypes.Bytes32 {
	publicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range publicKeys {
		publicKeyBytes := sender.Bytes() // Only x coordinate is used
		copy(publicKeysBytes[int32Key*i:int32Key*(i+1)], publicKeyBytes)
	}
	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < intMaxTypes.NumOfSenders; i++ {
		publicKeyBytes := dummyPublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(publicKeysBytes[int32Key*i:int32Key*(i+1)], publicKeyBytes[:])
	}

	publicKeysHash := crypto.Keccak256(publicKeysBytes) // TODO: Is this correct hash?

	var result intMaxTypes.Bytes32
	result.FromBytes(publicKeysHash)

	return result
}
