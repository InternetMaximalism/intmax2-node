package types

import (
	"encoding/binary"
	"errors"
	"intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	numPublicKeyBytes   = 32
	PublicKeySenderType = "PUBLIC_KEY"
	AccountIDSenderType = "ACCOUNT_ID"
)

type poseidonHashOut = goldenposeidon.PoseidonHashOut

// Sender represents an individual sender's details, including their public key, account ID,
// and a flag indicating if the sender has posted.
type Sender struct {
	PublicKey *accounts.PublicKey
	AccountID uint64
	IsSigned  bool
}

// BlockContent represents the content of a block, including sender details, transaction root,
// aggregated signature, and public key.
type BlockContent struct {
	// SenderType specifies whether senders are identified by PUBLIC_KEY or ACCOUNT_ID
	SenderType string

	// Senders is a list of senders in the block
	Senders []Sender

	// TxRoot is the root hash of the transactions in the block
	TxRoot poseidonHashOut

	// AggregatedSignature is the aggregated signature of the block
	AggregatedSignature *bn254.G2Affine

	// aggregatedPublicKey is the aggregated public key of the block
	AggregatedPublicKey *accounts.PublicKey

	MessagePoint *bn254.G2Affine
}

func NewBlockContent(senderType string, senders []Sender, txRoot poseidonHashOut, aggregatedSignature *bn254.G2Affine) *BlockContent {
	bc := new(BlockContent)
	bc.SenderType = senderType
	bc.Senders = make([]Sender, len(senders))
	copy(bc.Senders, senders)
	bc.TxRoot.Set(&txRoot)
	bc.AggregatedSignature = new(bn254.G2Affine).Set(aggregatedSignature)

	senderPublicKeys := make([]byte, len(bc.Senders)*numPublicKeyBytes)
	for i, sender := range bc.Senders {
		if sender.IsSigned {
			senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
		}
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeys)

	aggregatedPublicKey := accounts.NewPublicKey(new(bn254.G1Affine))
	for _, sender := range bc.Senders {
		if sender.IsSigned {
			aggregatedPublicKey.Pk.Add(aggregatedPublicKey.Pk, sender.PublicKey.WeightByHash(publicKeysHash).Pk)
		}
	}
	bc.AggregatedPublicKey = new(accounts.PublicKey).Set(aggregatedPublicKey)

	messagePoint := goldenposeidon.HashToG2(finite_field.BytesToFieldElementSlice(bc.TxRoot.Marshal()))
	bc.MessagePoint = &messagePoint

	return bc
}

func (bc *BlockContent) IsValid() error {
	if bc.SenderType != PublicKeySenderType && bc.SenderType != AccountIDSenderType {
		return errors.New("invalid sender type")
	}

	// Ensure there is at least one sender and no more than 128 senders
	if len(bc.Senders) == 0 {
		return errors.New("no senders")
	}
	if len(bc.Senders) > 128 {
		return errors.New("too many senders")
	}

	// Ensure public keys is sorted
	for i := 0; i < len(bc.Senders)-1; i++ {
		if bc.Senders[i+1].PublicKey.Pk.X.Cmp(&bc.Senders[i].PublicKey.Pk.X) > 0 {
			return errors.New("public keys are not sorted")
		}
	}

	switch bc.SenderType {
	case PublicKeySenderType:
		for _, sender := range bc.Senders {
			if sender.PublicKey == nil {
				return errors.New("invalid public key")
			}

			if sender.AccountID != 0 {
				return errors.New("invalid account ID: must be zero for PUBLIC_KEY sender type")
			}
		}
	case AccountIDSenderType:
		for _, sender := range bc.Senders {
			if sender.PublicKey == nil {
				return errors.New("invalid public key")
			}

			if sender.PublicKey.Pk.X.Cmp(new(fp.Element).SetOne()) != 0 && sender.AccountID == 0 {
				return errors.New("invalid account ID: must be non-zero for ACCOUNT_ID sender type")
			}
			if sender.PublicKey.Pk.X.Cmp(new(fp.Element).SetOne()) == 0 && sender.AccountID != 0 {
				return errors.New("invalid account ID: must be zero for default sender")
			}
		}
	default:
		return errors.New("invalid sender type")
	}

	// Check aggregated public key
	if bc.AggregatedPublicKey == nil {
		return errors.New("no aggregated public key")
	}
	senderPublicKeys := make([]byte, len(bc.Senders)*numPublicKeyBytes)
	for i, pk := range bc.Senders {
		if pk.IsSigned {
			senderPublicKey := pk.PublicKey.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
		}
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeys)
	aggregatedPublicKey := accounts.NewPublicKey(new(bn254.G1Affine))
	for _, sender := range bc.Senders {
		if sender.IsSigned {
			aggregatedPublicKey.Pk.Add(aggregatedPublicKey.Pk, sender.PublicKey.WeightByHash(publicKeysHash).Pk)
		}
	}

	if !aggregatedPublicKey.Equal(bc.AggregatedPublicKey) {
		return errors.New("invalid aggregated public key")
	}

	// Check aggregated signature
	if bc.AggregatedSignature == nil {
		return errors.New("no aggregated signature")
	}
	message := finite_field.BytesToFieldElementSlice(bc.TxRoot.Marshal())
	err := accounts.VerifySignature(bc.AggregatedSignature, bc.AggregatedPublicKey, message)
	if err != nil {
		return err
	}

	return nil
}

func (bc *BlockContent) Marshal() []byte {
	data := make([]byte, 0)

	if bc.SenderType == PublicKeySenderType {
		data = append(data, 0)
	} else {
		data = append(data, 1)
	}

	data = append(data, bc.TxRoot.Marshal()...)

	// TODO
	for _, sender := range bc.Senders {
		if sender.IsSigned {
			data = append(data, 1)
		} else {
			data = append(data, 0)
		}
	}

	numAccountIdBytes := 5
	senderAccountIDs := make([]byte, len(bc.Senders)*numAccountIdBytes)
	for i, pk := range bc.Senders {
		var senderAccountId []byte
		if bc.SenderType == AccountIDSenderType {
			publicKeyX := pk.PublicKey.Pk.X.Bytes() // TODO: Use account ID
			senderAccountId = publicKeyX[:5]
		} else {
			senderAccountId = []byte{0, 0, 0, 0, 0}
		}
		copy(senderAccountIDs[5*i:5*(i+1)], senderAccountId)
	}

	numPublicKeyBytes := 32
	senderPublicKeys := make([]byte, len(bc.Senders)*numPublicKeyBytes)
	for i, pk := range bc.Senders {
		senderPublicKey := pk.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
	}

	// messagePoint := goldenposeidon.HashToG2(finite_field.BytesToFieldElementSlice(bc.TxRoot.Marshal()))

	data = append(data, senderAccountIDs...)
	data = append(data, senderPublicKeys...)
	data = append(data, bc.AggregatedSignature.Marshal()...)

	return data
}

// The rollup's calldata consists of txRoot, messagePoint, aggregatedSignature, aggregatedPublicKey,
// accountIdsHash, senderPublicKeysHash, senderFlags and senderType.
// The size of the rollup data will be 32 + 128 + 128 + 64 + 32 + 32 + 16 + 1 = 433 bytes.
func (bc *BlockContent) Rollup() []byte {
	data := make([]byte, 0)

	data = append(data, bc.TxRoot.Marshal()...)

	data = append(data, bc.MessagePoint.Marshal()...)

	data = append(data, bc.AggregatedSignature.Marshal()...)

	data = append(data, bc.AggregatedPublicKey.Marshal()...)

	switch bc.SenderType {
	case PublicKeySenderType:
		senderPublicKeys := make([]byte, len(bc.Senders)*32)
		for i, pk := range bc.Senders {
			senderPublicKey := pk.PublicKey.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeys[32*i:32*(i+1)], senderPublicKey[:])
		}
		data = append(data, senderPublicKeys...)
	case AccountIDSenderType:
		senderAccountIDs := make([]byte, len(bc.Senders)*5)
		for i, pk := range bc.Senders {
			var senderAccountId []byte
			if bc.SenderType == AccountIDSenderType {
				publicKeyX := pk.PublicKey.Pk.X.Bytes() // TODO: Use account ID
				senderAccountId = publicKeyX[:5]
			} else {
				senderAccountId = []byte{0, 0, 0, 0, 0}
			}
			copy(senderAccountIDs[5*i:5*(i+1)], senderAccountId)
		}
		data = append(data, senderAccountIDs...)
	}

	numFlagBytes := (len(bc.Senders) + 7) / 8
	senderFlags := make([]byte, numFlagBytes)
	for i, pk := range bc.Senders {
		var isPosted uint8
		if pk.IsSigned {
			isPosted = 1
		} else {
			isPosted = 0
		}
		senderFlags[i/8] |= byte(isPosted << (uint(i) % 8))
	}
	data = append(data, senderFlags...)

	if bc.SenderType == PublicKeySenderType {
		data = append(data, 0)
	} else {
		data = append(data, 1)
	}

	// TODO: accountIDsHash *common.Hash

	// TODO: senderPublicKeysHash *common.Hash

	return data
}

func (bc *BlockContent) Hash() common.Hash {
	return crypto.Keccak256Hash(bc.Marshal())
}

type PostedBlock struct {
	// The previous block hash.
	PrevBlockHash common.Hash
	// The block number, which is the latest block number in the Rollup contract plus 1.
	BlockNumber uint32
	// The deposit root at the time of block posting (written in the Rollup contract).
	DepositRoot common.Hash
	// The hash value that the Block Builder must provide to the Rollup contract when posting a new block.
	ContentHash common.Hash
}

func NewPostedBlock(prevBlockHash, depositRoot common.Hash, blockNumber uint32, contentHash common.Hash) *PostedBlock {
	return &PostedBlock{
		PrevBlockHash: prevBlockHash,
		BlockNumber:   blockNumber,
		DepositRoot:   depositRoot,
		ContentHash:   contentHash,
	}
}

func (pb *PostedBlock) Marshal() []byte {
	data := make([]byte, 0)

	data = append(data, pb.PrevBlockHash.Bytes()...)
	blockNumberBytes := [4]byte{}
	binary.BigEndian.PutUint32(blockNumberBytes[:], pb.BlockNumber)
	data = append(data, blockNumberBytes[:]...)
	data = append(data, pb.DepositRoot.Bytes()...)
	data = append(data, pb.ContentHash.Bytes()...)

	return data
}

func (pb *PostedBlock) Hash() common.Hash {
	return crypto.Keccak256Hash(pb.Marshal())
}
