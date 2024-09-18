package post_withdrawal_request

import (
	"errors"
	"math/big"

	"github.com/prodadidb/go-validation"
)

const (
	base10 = 10
)

// ErrIsNegativeStr error: can not be non-negative.
var ErrIsNegativeStr = "can not be non-negative"

// ErrIsLessOrEqualZero error: can be more then zero.
var ErrIsLessOrEqualZero = errors.New("can be more then zero")

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCPostWithdrawalRequestInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.TransferData, validation.Required, input.validateTransferData()),
		validation.Field(&input.TransferMerkleProof, validation.Required, input.validateTransferMerkleProof()),
		validation.Field(&input.Transaction, validation.Required, input.validateTransaction()),
		validation.Field(&input.TxMerkleProof, validation.Required, input.validateTxMerkleProof()),
		validation.Field(&input.TransferHash, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
		validation.Field(&input.BlockHash, validation.Required),
		validation.Field(&input.EnoughBalanceProof, validation.Required, input.validateEnoughBalanceProof()),
	)
}

func (input *UCPostWithdrawalRequestInput) validateTransferData() validation.Rule {
	return validation.By(func(value interface{}) error {
		var isNil bool
		value, isNil = validation.Indirect(value)

		if isNil || validation.IsEmpty(value) {
			return ErrValueInvalid
		}

		transferData, ok := value.(UCPostWithdrawalRequestTransferDataInput)
		if !ok {
			return ErrValueInvalid
		}

		return validation.ValidateStruct(&transferData,
			validation.Field(&transferData.Recipient, validation.Required),
			validation.Field(&transferData.TokenIndex, validation.By(func(value interface{}) error {
				tokenIndex, okTokenIndex := value.(int64)
				if !okTokenIndex {
					return ErrValueInvalid
				}

				ti := int(tokenIndex)
				return validation.Validate(&ti, validation.Min(0).Error(ErrIsNegativeStr))
			})),
			validation.Field(&transferData.Amount, validation.Required, validation.By(func(value interface{}) error {
				amountV, okAmountV := value.(string)
				if !okAmountV {
					return ErrValueInvalid
				}

				amount, okAmount := new(big.Int).SetString(amountV, base10)
				if !okAmount {
					return ErrValueInvalid
				}

				if amount.Cmp(big.NewInt(0)) <= 0 {
					return ErrIsLessOrEqualZero
				}

				return nil
			})),
			validation.Field(&transferData.Salt, validation.Required),
		)
	})
}

func (input *UCPostWithdrawalRequestInput) validateTransferMerkleProof() validation.Rule {
	return validation.By(func(value interface{}) error {
		var isNil bool
		value, isNil = validation.Indirect(value)

		if isNil || validation.IsEmpty(value) {
			return ErrValueInvalid
		}

		transferMerkleProof, ok := value.(UCPostWithdrawalRequestTransferMerkleProofInput)
		if !ok {
			return ErrValueInvalid
		}

		return validation.ValidateStruct(&transferMerkleProof,
			validation.Field(&transferMerkleProof.Siblings, validation.Required, validation.Each(validation.Required)),
			validation.Field(&transferMerkleProof.Index, validation.Min(0).Error(ErrIsNegativeStr)),
		)
	})
}

func (input *UCPostWithdrawalRequestInput) validateTransaction() validation.Rule {
	return validation.By(func(value interface{}) error {
		var isNil bool
		value, isNil = validation.Indirect(value)

		if isNil || validation.IsEmpty(value) {
			return ErrValueInvalid
		}

		transaction, ok := value.(UCPostWithdrawalRequestTransactionInput)
		if !ok {
			return ErrValueInvalid
		}

		return validation.ValidateStruct(&transaction,
			validation.Field(&transaction.TransferTreeRoot, validation.Required),
			validation.Field(&transaction.Nonce, validation.Min(0).Error(ErrIsNegativeStr)),
		)
	})
}

func (input *UCPostWithdrawalRequestInput) validateTxMerkleProof() validation.Rule {
	return validation.By(func(value interface{}) error {
		var isNil bool
		value, isNil = validation.Indirect(value)

		if isNil || validation.IsEmpty(value) {
			return ErrValueInvalid
		}

		txMerkleProof, ok := value.(UCPostWithdrawalRequestTxMerkleProofInput)
		if !ok {
			return ErrValueInvalid
		}

		return validation.ValidateStruct(&txMerkleProof,
			validation.Field(&txMerkleProof.Siblings, validation.Required, validation.Each(validation.Required)),
			validation.Field(&txMerkleProof.Index, validation.Min(0).Error(ErrIsNegativeStr)),
		)
	})
}

func (input *UCPostWithdrawalRequestInput) validateEnoughBalanceProof() validation.Rule {
	return validation.By(func(value interface{}) error {
		var isNil bool
		value, isNil = validation.Indirect(value)

		if isNil || validation.IsEmpty(value) {
			return ErrValueInvalid
		}

		enoughBalanceProof, ok := value.(UCPostWithdrawalRequestEnoughBalanceProofInput)
		if !ok {
			return ErrValueInvalid
		}

		return validation.ValidateStruct(&enoughBalanceProof,
			validation.Field(&enoughBalanceProof.Proof, validation.Required),
			validation.Field(&enoughBalanceProof.PublicInputs, validation.Required),
		)
	})
}
