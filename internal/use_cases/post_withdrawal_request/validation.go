package post_withdrwal_request

import (
	"errors"
	"math/big"

	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/prodadidb/go-validation"
)

const (
	Base10 = 10
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCPostWithdrawalRequestInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.TransferData, validation.Required, validation.By(input.validateTransferData)),
		validation.Field(&input.TransferMerkleProof, validation.Required, validation.By(input.validateTransferMerkleProof)),
		validation.Field(&input.Transaction, validation.Required, validation.By(input.validateTransaction)),
		validation.Field(&input.TxMerkleProof, validation.Required, validation.By(input.validateTxMerkleProof)),
		validation.Field(&input.TransferHash, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
		validation.Field(&input.BlockHash, validation.Required),
		validation.Field(&input.EnoughBalanceProof, validation.Required, validation.By(input.validateEnoughBalanceProof)),
	)
}

func (input *UCPostWithdrawalRequestInput) validateTransferData(value interface{}) error {
	transferData, ok := value.(mDBApp.TransferData)
	if !ok {
		return ErrValueInvalid
	}

	if transferData.Recipient == "" {
		return errors.New("recipient must not be empty")
	}
	if transferData.TokenIndex < 0 {
		return errors.New("TokenIndex must be non-negative")
	}
	amount := new(big.Int)
	amount, ok = amount.SetString(transferData.Amount, Base10)
	if !ok {
		return errors.New("amount must be a valid number")
	}
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return errors.New("amount must be positive")
	}
	if transferData.Salt == "" {
		return errors.New("salt must not be empty")
	}
	return nil
}

func (input *UCPostWithdrawalRequestInput) validateTransferMerkleProof(value interface{}) error {
	proof, ok := value.(mDBApp.TransferMerkleProof)
	if !ok {
		return ErrValueInvalid
	}

	if len(proof.Siblings) == 0 {
		return errors.New("TransferMerkleProof Siblings must not be empty")
	}
	if proof.Index < 0 {
		return errors.New("TransferMerkleProof Index must be non-negative")
	}
	return nil
}

func (input *UCPostWithdrawalRequestInput) validateTransaction(value interface{}) error {
	transaction, ok := value.(mDBApp.Transaction)
	if !ok {
		return ErrValueInvalid
	}

	if transaction.TransferTreeRoot == "" {
		return errors.New("TransferTreeRoot must not be empty")
	}
	if transaction.Nonce < 0 {
		return errors.New("transaction Nonce must be non-negative")
	}
	return nil
}

func (input *UCPostWithdrawalRequestInput) validateTxMerkleProof(value interface{}) error {
	proof, ok := value.(mDBApp.TxMerkleProof)
	if !ok {
		return ErrValueInvalid
	}

	if len(proof.Siblings) == 0 {
		return errors.New("TxMerkleProof Siblings must not be empty")
	}
	if proof.Index < 0 {
		return errors.New("TxMerkleProof Index must be non-negative")
	}
	return nil
}

func (input *UCPostWithdrawalRequestInput) validateEnoughBalanceProof(value interface{}) error {
	proof, ok := value.(mDBApp.EnoughBalanceProof)
	if !ok {
		return ErrValueInvalid
	}

	if proof.Proof == "" {
		return errors.New("EnoughBalanceProof Proof must not be empty")
	}
	if proof.PublicInputs == "" {
		return errors.New("EnoughBalanceProof PublicInputs must not be empty")
	}
	return nil
}
