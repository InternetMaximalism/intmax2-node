package backup_balance

import (
	"errors"

	"github.com/prodadidb/go-validation"
)

const (
	Base10 = 10
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCGetBalancesInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Address, validation.Required),
	)
}

func (input *UCGetBackupBalancesInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required),
	)
}
