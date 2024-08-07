package backup_transaction

import (
	"errors"

	"github.com/prodadidb/go-validation"
)

const (
	Base10 = 10
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCPostBackupTransactionInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.EncryptedTx, validation.Required),
		validation.Field(&input.Sender, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
		validation.Field(&input.Signature, validation.Required),
	)
}
func (input *UCGetBackupTransactionInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required),
	)
}
