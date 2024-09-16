package block_post_service

import (
	"encoding/binary"
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/logger"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"sort"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const numOfSenders = intMaxTypes.NumOfSenders

type PostedBlock struct {
	// The previous block hash.
	PrevBlockHash common.Hash `json:"prevBlockHash"`
	// The block number, which is the latest block number in the Rollup contract plus 1.
	BlockNumber uint32 `json:"blockNumber"`
	// The deposit root at the time of block posting (written in the Rollup contract).
	DepositRoot common.Hash `json:"depositTreeRoot"`
	// The hash value that the Block Builder must provide to the Rollup contract when posting a new block.
	SignatureHash common.Hash `json:"signatureHash"`
}

func NewPostedBlock(prevBlockHash, depositRoot common.Hash, blockNumber uint32, signatureHash common.Hash) *PostedBlock {
	return &PostedBlock{
		PrevBlockHash: prevBlockHash,
		BlockNumber:   blockNumber,
		DepositRoot:   depositRoot,
		SignatureHash: signatureHash,
	}
}

func (pb *PostedBlock) Set(other *PostedBlock) *PostedBlock {
	pb.PrevBlockHash = other.PrevBlockHash
	copy(pb.PrevBlockHash[:], other.PrevBlockHash[:])
	pb.DepositRoot = other.DepositRoot
	copy(pb.DepositRoot[:], other.DepositRoot[:])
	pb.SignatureHash = other.SignatureHash
	copy(pb.SignatureHash[:], other.SignatureHash[:])
	pb.BlockNumber = other.BlockNumber

	return pb
}

func (pb *PostedBlock) Equals(other *PostedBlock) bool {
	return pb.PrevBlockHash == other.PrevBlockHash &&
		pb.DepositRoot == other.DepositRoot &&
		pb.SignatureHash == other.SignatureHash &&
		pb.BlockNumber == other.BlockNumber
}

func (pb *PostedBlock) Genesis() *PostedBlock {
	depositTree, err := intMaxTree.NewDepositTree(intMaxTree.DEPOSIT_TREE_HEIGHT)
	if err != nil {
		panic(err)
	}

	depositTreeRoot, _, _ := depositTree.GetCurrentRootCountAndSiblings()

	return NewPostedBlock(common.Hash{}, depositTreeRoot, 0, common.Hash{})
}

func (pb *PostedBlock) Marshal() []byte {
	const int4Key = 4

	data := make([]byte, 0)

	data = append(data, pb.PrevBlockHash.Bytes()...)
	data = append(data, pb.DepositRoot.Bytes()...)
	data = append(data, pb.SignatureHash.Bytes()...)
	blockNumberBytes := [int4Key]byte{}
	binary.BigEndian.PutUint32(blockNumberBytes[:], pb.BlockNumber)
	data = append(data, blockNumberBytes[:]...)

	return data
}

func (pb *PostedBlock) Uint32Slice() []uint32 {
	var buf []uint32
	buf = append(buf, intMaxTypes.CommonHashToUint32Slice(pb.PrevBlockHash)...)
	buf = append(buf, intMaxTypes.CommonHashToUint32Slice(pb.DepositRoot)...)
	buf = append(buf, intMaxTypes.CommonHashToUint32Slice(pb.SignatureHash)...)
	buf = append(buf, pb.BlockNumber)

	return buf
}

func (pb *PostedBlock) Hash() common.Hash {
	return crypto.Keccak256Hash(intMaxTypes.Uint32SliceToBytes(pb.Uint32Slice()))
}

// MakeRegistrationBlock creates a block content for registration block.
// txRoot - root of the transaction tree.
// senderPublicKeys - list of public keys for each sender.
// signatures - list of signatures for each sender. Empty string means no signature.
func MakeRegistrationBlock(
	txRoot intMaxTypes.PoseidonHashOut,
	senderPublicKeys []*intMaxAcc.PublicKey,
	signatures []string,
) (*intMaxTypes.BlockContent, error) {
	if len(senderPublicKeys) != len(signatures) {
		return nil, errors.New("length of senderPublicKeys, accountIDs, and signatures must be equal")
	}

	// Sort by x-coordinate of public key
	sort.Slice(senderPublicKeys, func(i, j int) bool {
		return senderPublicKeys[i].Pk.X.Cmp(&senderPublicKeys[j].Pk.X) > 0
	})

	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, publicKey := range senderPublicKeys {
		if publicKey == nil {
			return nil, errors.New("publicKey must not be nil")
		}

		senders[i] = intMaxTypes.Sender{
			PublicKey: publicKey,
			AccountID: 0,
			IsSigned:  signatures[i] != "",
		}
	}

	defaultSender := intMaxTypes.NewDummySender()
	for i := len(senderPublicKeys); i < len(senders); i++ {
		senders[i] = defaultSender
	}

	const numPublicKeyBytes = intMaxTypes.NumPublicKeyBytes

	senderPublicKeysBytes := make([]byte, len(senders)*numPublicKeyBytes)
	for i, sender := range senders {
		senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[numPublicKeyBytes*i:numPublicKeyBytes*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for _, sender := range senders {
		if sender.IsSigned {
			weightedPublicKey := sender.PublicKey.WeightByHash(publicKeysHash)
			aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
		}
	}

	aggregatedSignature := new(bn254.G2Affine)
	for i, sender := range senders {
		if senders[i].IsSigned {
			if sender.IsSigned {
				signature := new(bn254.G2Affine)
				signatureBytes, err := hexutil.Decode(signatures[i])
				if err != nil {
					fmt.Printf("Failed to decode signature: %s\n", signatures[i])
					continue
				}
				err = signature.Unmarshal(signatureBytes)
				if err != nil {
					fmt.Printf("Failed to unmarshal signature: %s\n", signatures[i])
					continue
				}

				err = VerifyTxTreeSignature(
					signatureBytes, sender.PublicKey, txRoot.Marshal(), senderPublicKeys,
				)
				if err != nil {
					fmt.Printf("Failed to verify signature: %s\n", signatures[i])
					continue
				}

				aggregatedSignature.Add(aggregatedSignature, signature)
			}
		}
	}

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.PublicKeySenderType,
		senders,
		txRoot,
		aggregatedSignature,
	)
	err := blockContent.IsValid()
	if err != nil {
		return nil, errors.Join(ErrInvalidRegistrationBlockContent, err)
	}

	return blockContent, nil
}

// MakeNonRegistrationBlock creates a block content for non-registration block.
// txRoot - root of the transaction tree.
// accountIDs - list of account IDs for each sender.
// senderPublicKeys - list of public keys for each sender.
// signatures - list of signatures for each sender. Empty string means no signature.
func MakeNonRegistrationBlock(
	txRoot intMaxTypes.PoseidonHashOut,
	accountIDs []uint64,
	senderPublicKeys []*intMaxAcc.PublicKey,
	signatures []string,
) (*intMaxTypes.BlockContent, error) {
	if len(senderPublicKeys) != len(signatures) || len(senderPublicKeys) != len(accountIDs) {
		return nil, errors.New("length of senderPublicKeys, accountIDs, and signatures must be equal")
	}

	// Sort by x-coordinate of public key
	sort.Slice(senderPublicKeys, func(i, j int) bool {
		return senderPublicKeys[i].Pk.X.Cmp(&senderPublicKeys[j].Pk.X) > 0
	})

	const maxAccountIDBits = 40

	senders := make([]intMaxTypes.Sender, numOfSenders)
	for i, publicKey := range senderPublicKeys {
		if accountIDs[i] == 0 {
			return nil, errors.New("accountID must be greater than 0")
		}
		if accountIDs[i] > uint64(1)<<maxAccountIDBits {
			return nil, fmt.Errorf("accountID must be less than or equal to 2^%d", maxAccountIDBits)
		}
		if publicKey == nil {
			return nil, errors.New("publicKey must not be nil")
		}

		senders[i] = intMaxTypes.Sender{
			PublicKey: publicKey,
			AccountID: accountIDs[i],
			IsSigned:  signatures[i] != "",
		}
	}

	defaultPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(senderPublicKeys); i < len(senders); i++ {
		senders[i] = intMaxTypes.Sender{
			PublicKey: defaultPublicKey,
			AccountID: 1,
			IsSigned:  false,
		}
	}

	const numPublicKeyBytes = intMaxTypes.NumPublicKeyBytes

	senderPublicKeysBytes := make([]byte, len(senders)*numPublicKeyBytes)
	for i, sender := range senders {
		senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[numPublicKeyBytes*i:numPublicKeyBytes*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)
	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for _, sender := range senders {
		if sender.IsSigned {
			weightedPublicKey := sender.PublicKey.WeightByHash(publicKeysHash)
			aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
		}
	}

	aggregatedSignature := new(bn254.G2Affine)
	for i, sender := range senders {
		if senders[i].IsSigned {
			if sender.IsSigned {
				signature := new(bn254.G2Affine)
				signatureBytes, err := hexutil.Decode(signatures[i])
				if err != nil {
					fmt.Printf("Failed to decode signature: %s\n", signatures[i])
					continue
				}
				err = signature.Unmarshal(signatureBytes)
				if err != nil {
					fmt.Printf("Failed to unmarshal signature: %s\n", signatures[i])
					continue
				}

				err = VerifyTxTreeSignature(
					signatureBytes, sender.PublicKey, txRoot.Marshal(), senderPublicKeys,
				)
				if err != nil {
					fmt.Printf("Failed to verify signature: %s\n", signatures[i])
					continue
				}

				aggregatedSignature.Add(aggregatedSignature, signature)
			}
		}
	}

	blockContent := intMaxTypes.NewBlockContent(
		intMaxTypes.AccountIDSenderType,
		senders,
		txRoot,
		aggregatedSignature,
	)
	err := blockContent.IsValid()
	if err != nil {
		return nil, errors.Join(ErrInvalidNonRegistrationBlockContent, err)
	}

	return blockContent, nil
}

func VerifyTxTreeSignature(signatureBytes []byte, sender *intMaxAcc.PublicKey, txTreeRootBytes []byte, senderPublicKeys []*intMaxAcc.PublicKey) error {
	const int32Key = 32

	if len(senderPublicKeys) == 0 {
		return ErrInvalidSendersLength
	}
	if len(senderPublicKeys) > intMaxTypes.NumOfSenders {
		return ErrTooManySenderPublicKeys
	}

	// Sort by x-coordinate of public key
	sort.Slice(senderPublicKeys, func(i, j int) bool {
		return senderPublicKeys[i].Pk.X.Cmp(&senderPublicKeys[j].Pk.X) > 0
	})

	senderPublicKeysBytes := make([]byte, intMaxTypes.NumOfSenders*intMaxTypes.NumPublicKeyBytes)
	for i, sender := range senderPublicKeys {
		senderPublicKey := sender.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+1)], senderPublicKey[:])
	}
	defaultPublicKey := intMaxAcc.NewDummyPublicKey()
	for i := len(senderPublicKeys); i < intMaxTypes.NumOfSenders; i++ {
		senderPublicKey := defaultPublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeysBytes[int32Key*i:int32Key*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeysBytes)

	messagePoint := finite_field.BytesToFieldElementSlice(txTreeRootBytes)

	signature := new(bn254.G2Affine)
	err := signature.Unmarshal(signatureBytes)
	if err != nil {
		return errors.Join(ErrUnmarshalSignatureFail, err)
	}

	senderWithWeight := sender.WeightByHash(publicKeysHash)
	err = intMaxAcc.VerifySignature(signature, senderWithWeight, messagePoint)
	if err != nil {
		return errors.Join(ErrInvalidSignature, err)
	}

	return nil
}

func UpdateEventBlockNumber(db SQLDriverApp, log logger.Logger, eventName string, blockNumber uint64) error {
	updatedEvent, err := db.UpsertEventBlockNumber(eventName, blockNumber)
	if err != nil {
		return err
	}
	log.Infof("Updated %s block number to %d", eventName, updatedEvent.LastProcessedBlockNumber)
	return nil
}
