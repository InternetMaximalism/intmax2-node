package block_signature

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/worker"
	"sort"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prodadidb/go-validation"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

// ErrTransactionHashNotFound error: the transaction hash not found.
var ErrTransactionHashNotFound = errors.New("the transaction hash not found")

// ErrTxTreeNotBuild error: the tx tree not build.
var ErrTxTreeNotBuild = errors.New("the tx tree not build")

// ErrTxTreeSignatureCollectionComplete error: signature collection for tx tree completed.
var ErrTxTreeSignatureCollectionComplete = errors.New("signature collection for tx tree completed")

func (input *UCBlockSignatureInput) Valid(w Worker) (err error) {
	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required, input.isSender(func() *intMaxAcc.PublicKey {
			if input.DecodeSender == nil {
				input.DecodeSender = &intMaxAcc.PublicKey{}
			}
			return input.DecodeSender
		}())),
		validation.Field(&input.TxHash, validation.Required, input.isHexDecode(), input.isExistsTxHash(w)),
		validation.Field(&input.EnoughBalanceProof, validation.Required, input.isEnoughBalanceProof()),
		validation.Field(&input.Signature, validation.Required, validation.By(func(value interface{}) (err error) {
			v, ok := value.(string)
			if !ok {
				return ErrValueInvalid
			}

			txTreeRootBytes := input.TxTree.RootHash.Marshal()

			var publicKey *intMaxAcc.PublicKey
			publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(input.Sender)
			if err != nil {
				return ErrValueInvalid
			}

			var sb []byte
			sb, err = hexutil.Decode(v)
			if err != nil {
				return ErrValueInvalid
			}

			// TODO: Include all public keys contained in the tx tree.
			senderPublicKeys := make([]*intMaxAcc.PublicKey, 1)
			senderPublicKeys[0] = publicKey

			// Verify signature.
			err = VerifyTxTreeSignature(sb, publicKey, txTreeRootBytes, senderPublicKeys)
			if err != nil {
				fmt.Printf("VerifySignature error: %v\n", err)
				// TODO: error handling: return ErrValueInvalid
			}

			return nil
		})),
	)
}

func (input *UCBlockSignatureInput) isSender(pbKey *intMaxAcc.PublicKey) validation.Rule {
	return validation.By(func(value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		publicKey, err := intMaxAcc.NewPublicKeyFromAddressHex(v)
		if err != nil {
			return ErrValueInvalid
		}

		if pbKey != nil {
			*pbKey = *publicKey
		}

		return nil
	})
}

func (input *UCBlockSignatureInput) isHexDecode() validation.Rule {
	return validation.By(func(value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		_, err := hexutil.Decode(v)
		if err != nil {
			return ErrValueInvalid
		}

		return nil
	})
}

func (input *UCBlockSignatureInput) isExistsTxHash(w Worker) validation.Rule {
	return validation.By(func(value interface{}) (err error) {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		var info *worker.TransactionHashesWithSenderAndFile
		info, err = w.TrHash(v)
		if err != nil && errors.Is(err, worker.ErrTransactionHashNotFound) ||
			!strings.EqualFold(info.Sender, input.Sender) {
			return ErrTransactionHashNotFound
		}

		input.TxInfo = info

		var txTree *worker.TxTree
		txTree, err = w.TxTreeByAvailableFile(info)
		if err != nil {
			switch {
			case errors.Is(err, worker.ErrTxTreeByAvailableFileFail):
				return ErrTransactionHashNotFound
			case errors.Is(err, worker.ErrTxTreeNotFound):
				return ErrTxTreeNotBuild
			case errors.Is(err, worker.ErrTxTreeSignatureCollectionComplete):
				return ErrTxTreeSignatureCollectionComplete
			default:
				return ErrValueInvalid
			}
		}

		input.TxTree = txTree

		return nil
	})
}

func (input *UCBlockSignatureInput) isEnoughBalanceProof() validation.Rule {
	return validation.By(func(value interface{}) error {
		var isNil bool
		value, isNil = validation.Indirect(value)

		if isNil || validation.IsEmpty(value) {
			return nil
		}

		ebp, ok := value.(EnoughBalanceProofInput)
		if !ok {
			return ErrValueInvalid
		}

		return validation.ValidateStruct(&ebp,
			validation.Field(&ebp.TransferStepProof, validation.Required, input.isPlonky2Proof()),
			validation.Field(&ebp.PrevBalanceProof, validation.Required, input.isPlonky2Proof()),
		)
	})
}

func (input *UCBlockSignatureInput) isPlonky2Proof() validation.Rule {
	return validation.By(func(value interface{}) error {
		var isNil bool
		value, isNil = validation.Indirect(value)

		if isNil || validation.IsEmpty(value) {
			return nil
		}

		p2p, ok := value.(Plonky2Proof)
		if !ok {
			return ErrValueInvalid
		}

		return validation.ValidateStruct(&p2p,
			validation.Field(&p2p.PublicInputs, validation.Required),
			validation.Field(&p2p.Proof, validation.Required),
		)
	})
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
