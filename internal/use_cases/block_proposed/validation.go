package block_proposed

import (
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/worker"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prodadidb/go-validation"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

// ErrTransfersHashNotFound error: the transfers hash not found.
var ErrTransfersHashNotFound = errors.New("the transfers hash not found")

// ErrTxTreeNotBuild error: the tx tree not build.
var ErrTxTreeNotBuild = errors.New("the tx tree not build")

// ErrTxTreeSignatureCollectionComplete error: signature collection for tx tree completed.
var ErrTxTreeSignatureCollectionComplete = errors.New("signature collection for tx tree completed")

func (input *UCBlockProposedInput) Valid(w Worker) error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required, input.isSender(func() *intMaxAcc.PublicKey {
			if input.DecodeSender == nil {
				input.DecodeSender = &intMaxAcc.PublicKey{}
			}
			return input.DecodeSender
		}())),
		validation.Field(&input.TxHash,
			validation.Required, input.isHexDecode(), input.isExistsTxHash(w),
		),
	)
}

func (input *UCBlockProposedInput) isSender(pbKey *intMaxAcc.PublicKey) validation.Rule {
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

func (input *UCBlockProposedInput) isHexDecode() validation.Rule {
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

func (input *UCBlockProposedInput) isExistsTxHash(w Worker) validation.Rule {
	return validation.By(func(value interface{}) (err error) {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		var info *worker.TransferHashesWithSenderAndFile
		info, err = w.TrHash(v)
		if err != nil && errors.Is(err, worker.ErrTransfersHashNotFound) ||
			!strings.EqualFold(info.Sender, input.Sender) {
			return ErrTransfersHashNotFound
		}

		var txTree *worker.TxTree
		txTree, err = w.TxTreeByAvailableFile(info)
		if err != nil {
			switch {
			case errors.Is(err, worker.ErrTxTreeByAvailableFileFail):
				return ErrTransfersHashNotFound
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
