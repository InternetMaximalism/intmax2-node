package verify_deposit_confirmation

import (
	"errors"

	"github.com/prodadidb/go-validation"
)

const (
	Base10 = 10
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCGetVerifyDepositConfirmationInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.DepositId, validation.Required),
	)
}
