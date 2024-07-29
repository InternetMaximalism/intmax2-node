package block_signature

import (
	"errors"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/worker"
	"strings"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

			// TODO: senderPublicKey with weight

			// Verify signature.
			err = VerifyTxTreeSignature(sb, publicKey, txTreeRootBytes)
			if err != nil {
				// TODO: error handling
				fmt.Printf("VerifySignature error: %v\n", err)
				// return ErrValueInvalid
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

func VerifyTxTreeSignature(signatureBytes []byte, sender *intMaxAcc.PublicKey, txTreeRootBytes []byte) error {
	messagePoint := finite_field.BytesToFieldElementSlice(txTreeRootBytes)

	signature := new(bn254.G2Affine)
	err := signature.Unmarshal(signatureBytes)
	if err != nil {
		return errors.Join(ErrUnmarshalSignatureFail, err)
	}

	err = intMaxAcc.VerifySignature(signature, sender, messagePoint)
	if err != nil {
		return errors.Join(ErrInvalidSignature, err)
	}

	return nil
}
