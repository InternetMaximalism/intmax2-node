package types

import (
	"intmax2-node/internal/accounts"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prodadidb/go-validation"
)

const (
	NumPublicKeyBytes   = 32
	PublicKeySenderType = "PUBLIC_KEY"

	NumAccountIDBytes   = 5
	AccountIDSenderType = "ACCOUNT_ID"
)

type PoseidonHashOut = goldenposeidon.PoseidonHashOut

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
	TxTreeRoot PoseidonHashOut

	// AggregatedSignature is the aggregated signature of the block
	AggregatedSignature *bn254.G2Affine

	// aggregatedPublicKey is the aggregated public key of the block
	AggregatedPublicKey *accounts.PublicKey

	MessagePoint *bn254.G2Affine
}

func NewBlockContent(
	senderType string,
	senders []Sender,
	txTreeRoot PoseidonHashOut,
	aggregatedSignature *bn254.G2Affine,
) *BlockContent {
	var bc BlockContent
	bc.SenderType = senderType
	bc.Senders = make([]Sender, len(senders))
	copy(bc.Senders, senders)
	bc.TxTreeRoot.Set(&txTreeRoot)
	bc.AggregatedSignature = new(bn254.G2Affine).Set(aggregatedSignature)

	defaultPublicKey := accounts.NewDummyPublicKey()

	const numOfSenders = 128
	senderPublicKeys := make([]byte, numOfSenders*NumPublicKeyBytes)
	for i, sender := range bc.Senders {
		senderPublicKey := sender.PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[NumPublicKeyBytes*i:NumPublicKeyBytes*(i+1)], senderPublicKey[:])
	}
	for i := len(bc.Senders); i < numOfSenders; i++ {
		senderPublicKey := defaultPublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[NumPublicKeyBytes*i:NumPublicKeyBytes*(i+1)], senderPublicKey[:])
	}

	publicKeysHash := crypto.Keccak256(senderPublicKeys)

	aggregatedPublicKey := new(accounts.PublicKey)
	for _, sender := range bc.Senders {
		if sender.IsSigned {
			aggregatedPublicKey.Add(aggregatedPublicKey, sender.PublicKey.WeightByHash(publicKeysHash))
		}
	}
	bc.AggregatedPublicKey = new(accounts.PublicKey).Set(aggregatedPublicKey)

	messagePoint := goldenposeidon.HashToG2(finite_field.BytesToFieldElementSlice(bc.TxTreeRoot.Marshal()))
	bc.MessagePoint = &messagePoint

	return &bc
}

func (bc *BlockContent) IsValid() error {
	const (
		int0Key   = 0
		int1Key   = 1
		int128Key = 128
	)

	return validation.ValidateStruct(bc,
		validation.Field(&bc.SenderType,
			validation.Required.Error(ErrBlockContentSenderTypeInvalid.Error()),
			validation.In(PublicKeySenderType, AccountIDSenderType).Error(ErrBlockContentSenderTypeInvalid.Error())),
		validation.Field(&bc.Senders,
			validation.Required.Error(ErrBlockContentSendersEmpty.Error()),
			validation.By(func(value interface{}) error {
				v, ok := value.([]Sender)
				if !ok {
					return ErrValueInvalid
				}

				if len(v) > int128Key {
					return ErrBlockContentManySenders
				}

				for i := int0Key; i < len(v)-int1Key; i++ {
					if v[i+int1Key].PublicKey.Pk.X.Cmp(&v[i].PublicKey.Pk.X) > int0Key {
						return ErrBlockContentPublicKeyNotSorted
					}
				}

				return nil
			}),
			validation.Each(validation.Required, validation.By(func(value interface{}) error {
				v, ok := value.(Sender)
				if !ok {
					return ErrValueInvalid
				}

				switch bc.SenderType {
				case PublicKeySenderType:
					if v.PublicKey == nil {
						return ErrBlockContentPublicKeyInvalid
					}

					if v.AccountID != int0Key {
						return ErrBlockContentAccIDForPubKeyInvalid
					}
				case AccountIDSenderType:
					if v.PublicKey == nil {
						return ErrBlockContentPublicKeyInvalid
					}

					if v.AccountID == int0Key && v.PublicKey.Pk.X.Cmp(new(fp.Element).SetOne()) != int0Key {
						return ErrBlockContentAccIDForAccIDEmpty
					}
					if v.AccountID != int0Key && v.PublicKey.Pk.X.Cmp(new(fp.Element).SetOne()) == int0Key {
						return ErrBlockContentAccIDForDefAccNotEmpty
					}
				}

				return nil
			}))),
		validation.Field(&bc.AggregatedPublicKey,
			validation.By(func(value interface{}) error {
				var isNil bool
				value, isNil = validation.Indirect(value)
				if isNil || validation.IsEmpty(value) {
					return ErrBlockContentAggPubKeyEmpty
				}

				defaultPublicKey := intMaxAcc.NewDummyPublicKey()

				const numOfSenders = 128
				senderPublicKeys := make([]byte, numOfSenders*NumPublicKeyBytes)
				for key := range bc.Senders {
					senderPublicKey := bc.Senders[key].PublicKey.Pk.X.Bytes() // Only x coordinate is used
					copy(
						senderPublicKeys[NumPublicKeyBytes*key:NumPublicKeyBytes*(key+int1Key)],
						senderPublicKey[:],
					)
				}
				for i := len(bc.Senders); i < numOfSenders; i++ {
					senderPublicKey := defaultPublicKey.Pk.X.Bytes() // Only x coordinate is used
					copy(senderPublicKeys[NumPublicKeyBytes*i:NumPublicKeyBytes*(i+1)], senderPublicKey[:])
				}

				publicKeysHash := crypto.Keccak256(senderPublicKeys)
				aggregatedPublicKey := new(accounts.PublicKey)
				for key := range bc.Senders {
					if bc.Senders[key].IsSigned {
						aggregatedPublicKey.Add(
							aggregatedPublicKey,
							bc.Senders[key].PublicKey.WeightByHash(publicKeysHash),
						)
					}
				}

				if !aggregatedPublicKey.Equal(bc.AggregatedPublicKey) {
					return ErrBlockContentAggPubKeyInvalid
				}

				return nil
			}),
		),
		validation.Field(&bc.AggregatedSignature,
			validation.By(func(value interface{}) error {
				var isNil bool
				value, isNil = validation.Indirect(value)
				if isNil || validation.IsEmpty(value) {
					return ErrBlockContentAggSignEmpty
				}

				message := finite_field.BytesToFieldElementSlice(bc.TxTreeRoot.Marshal())
				err := accounts.VerifySignature(bc.AggregatedSignature, bc.AggregatedPublicKey, message)
				if err != nil {
					return err
				}

				return nil
			}),
		),
	)
}

func (bc *BlockContent) Marshal() []byte {
	const (
		int0Key = 0
		int1Key = 1
	)

	var data []byte
	if bc.SenderType == PublicKeySenderType {
		data = append(data, int0Key)
	} else {
		data = append(data, int1Key)
	}
	data = append(data, bc.TxTreeRoot.Marshal()...)

	// TODO: need check
	for key := range bc.Senders {
		if bc.Senders[key].IsSigned {
			data = append(data, int1Key)
		} else {
			data = append(data, int0Key)
		}
	}

	senderAccountIDs := make([]byte, len(bc.Senders)*NumAccountIDBytes)
	for key := range bc.Senders {
		var senderAccountId []byte
		if bc.SenderType == AccountIDSenderType {
			publicKeyX := bc.Senders[key].PublicKey.Pk.X.Bytes() // TODO: Use account ID
			senderAccountId = publicKeyX[:NumAccountIDBytes]
		} else {
			senderAccountId = []byte{int0Key, int0Key, int0Key, int0Key, int0Key}
		}
		copy(senderAccountIDs[NumAccountIDBytes*key:NumAccountIDBytes*(key+int1Key)], senderAccountId)
	}

	senderPublicKeys := make([]byte, len(bc.Senders)*NumPublicKeyBytes)
	for key := range bc.Senders {
		senderPublicKey := bc.Senders[key].PublicKey.Pk.X.Bytes() // Only x coordinate is used
		copy(senderPublicKeys[NumPublicKeyBytes*key:NumPublicKeyBytes*(key+int1Key)], senderPublicKey[:])
	}

	data = append(data, senderAccountIDs...)
	data = append(data, senderPublicKeys...)
	data = append(data, bc.AggregatedSignature.Marshal()...)

	return data
}

// Rollup is
// txRoot, messagePoint, aggregatedSignature, aggregatedPublicKey,
// accountIdsHash, senderPublicKeysHash, senderFlags, senderType
// The size of the Rollup data will be 32 + 128 + 128 + 64 + 32 + 32 + 16 + 1 = 433 bytes
func (bc *BlockContent) Rollup() []byte {
	const (
		int0Key = 0
		int1Key = 1
		int8Key = 8
	)

	var data []byte
	data = append(data, bc.TxTreeRoot.Marshal()...)
	data = append(data, bc.MessagePoint.Marshal()...)
	data = append(data, bc.AggregatedSignature.Marshal()...)
	data = append(data, bc.AggregatedPublicKey.Marshal()...)

	switch bc.SenderType {
	case PublicKeySenderType:
		senderPublicKeys := make([]byte, len(bc.Senders)*NumPublicKeyBytes)
		for key := range bc.Senders {
			senderPublicKey := bc.Senders[key].PublicKey.Pk.X.Bytes() // Only x coordinate is used
			copy(senderPublicKeys[NumPublicKeyBytes*key:NumPublicKeyBytes*(key+int1Key)], senderPublicKey[:])
		}
		data = append(data, senderPublicKeys...)
	case AccountIDSenderType:
		senderAccountIDs := make([]byte, len(bc.Senders)*NumAccountIDBytes)
		for key := range bc.Senders {
			var senderAccountId []byte
			if bc.SenderType == AccountIDSenderType {
				publicKeyX := bc.Senders[key].PublicKey.Pk.X.Bytes() // TODO: Use account ID
				senderAccountId = publicKeyX[:NumAccountIDBytes]
			} else {
				senderAccountId = []byte{int0Key, int0Key, int0Key, int0Key, int0Key}
			}
			copy(senderAccountIDs[NumAccountIDBytes*key:NumAccountIDBytes*(key+int1Key)], senderAccountId)
		}
		data = append(data, senderAccountIDs...)
	}

	senderFlags := make([]byte, len(bc.Senders)/int8Key)
	for key := range bc.Senders {
		var isPosted uint8
		if bc.Senders[key].IsSigned {
			isPosted = int1Key
		} else {
			isPosted = int0Key
		}
		senderFlags[key/int8Key] |= isPosted << (uint(key) % int8Key)
	}
	data = append(data, senderFlags...)

	if bc.SenderType == PublicKeySenderType {
		data = append(data, int0Key)
	} else {
		data = append(data, int1Key)
	}

	return data
}

func (bc *BlockContent) Hash() common.Hash {
	return crypto.Keccak256Hash(bc.Marshal())
}
