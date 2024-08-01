package backup_deposit

import (
	"errors"

	"github.com/prodadidb/go-validation"
)

const (
	Base10 = 10
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCPostBackupDepositInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.EncryptedDeposit, validation.Required),
		validation.Field(&input.Recipient, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
	)
}
