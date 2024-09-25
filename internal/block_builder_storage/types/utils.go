package types

import (
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/ffg"
)

const (
	Int4Key  = 4
	Int5Key  = 5
	Int8Key  = 8
	Int16Key = 16
	Int32Key = 32

	Base10 = 10
)

// NewBlockHashTree is a Merkle tree that includes the genesis block in the 0th leaf from the beginning.
func NewBlockHashTree(height uint8) (*intMaxTree.BlockHashTree, error) {
	genesisBlock := new(block_post_service.PostedBlock).Genesis()
	genesisBlockHash := intMaxTree.NewBlockHashLeaf(genesisBlock.Hash())
	initialLeaves := []*intMaxTree.BlockHashLeaf{genesisBlockHash}

	return intMaxTree.NewBlockHashTreeWithInitialLeaves(height, initialLeaves)
}

func GetBitFromUint32Slice(limbs []uint32, i int) bool {
	if i >= len(limbs)*Int32Key {
		panic("out of index")
	}

	return (limbs[i/Int32Key]>>(Int32Key-1-i%Int32Key))&1 == 1
}

func EffectiveBits(n uint) uint32 {
	if n == 0 {
		return 0
	}

	bits := uint32(0)
	for n > 0 {
		n >>= 1
		bits++
	}

	return bits
}

func FieldElementSliceToUint32Slice(value []ffg.Element) []uint32 {
	v := make([]uint32, len(value))
	for i, x := range value {
		y := x.ToUint64Regular()
		if y >= uint64(1)<<Int32Key {
			panic("overflow")
		}
		v[i] = uint32(y)
	}

	return v
}

func GetPublicKeysHash(publicKeys []intMaxTypes.Uint256) intMaxTypes.Bytes32 {
	publicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range publicKeys {
		publicKeyBytes := sender.Bytes() // Only x coordinate is used
		copy(publicKeysBytes[Int32Key*i:Int32Key*(i+1)], publicKeyBytes)
	}
	dummyPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(publicKeys); i < intMaxTypes.NumOfSenders; i++ {
		publicKeyBytes := dummyPublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(publicKeysBytes[Int32Key*i:Int32Key*(i+1)], publicKeyBytes[:])
	}

	publicKeysHash := crypto.Keccak256(publicKeysBytes) // TODO: Is this correct hash?

	var result intMaxTypes.Bytes32
	result.FromBytes(publicKeysHash)

	return result
}

func GetPublicKeyCommitment(publicKeys []intMaxTypes.Uint256) *intMaxGP.PoseidonHashOut {
	publicKeyFlattened := make([]ffg.Element, 0)
	for _, publicKey := range publicKeys {
		publicKeyFlattened = append(publicKeyFlattened, publicKey.ToFieldElementSlice()...)
	}

	return intMaxGP.HashNoPad(publicKeyFlattened)
}

func GetSenderTreeRoot(publicKeys []intMaxTypes.Uint256, senderFlag intMaxTypes.Bytes16) *intMaxGP.PoseidonHashOut {
	if len(publicKeys) != intMaxTypes.NumOfSenders {
		panic("public keys length should be equal to number of senders")
	}

	senderLeafHashes := make([]*intMaxGP.PoseidonHashOut, len(publicKeys))
	for i, publicKey := range publicKeys {
		isValid := GetBitFromUint32Slice(senderFlag[:], i)
		senderLeaf := SenderLeaf{Sender: publicKey.BigInt(), IsValid: isValid}
		senderLeafHashes[i] = senderLeaf.Hash()
	}

	zeroHash := new(intMaxGP.PoseidonHashOut).SetZero()
	senderTree, err := intMaxTree.NewPoseidonIncrementalMerkleTree(intMaxTree.TX_TREE_HEIGHT, senderLeafHashes, zeroHash)
	if err != nil {
		panic(err)
	}

	root, _, _ := senderTree.GetCurrentRootCountAndSiblings()

	return &root
}

func GetAccountIDsHash(accountIDs []uint64) intMaxTypes.Bytes32 {
	accountIDsPacked := new(AccountIdPacked).Pack(accountIDs)

	return accountIDsPacked.Hash()
}
