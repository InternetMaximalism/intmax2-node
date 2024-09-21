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

func (input *UCPostBackupBalanceInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.User, validation.Required),
		validation.Field(&input.EncryptedBalanceProof, validation.Required),
		validation.Field(&input.EncryptedBalanceData, validation.Required),
		// validation.Field(&input.EncryptedTxs, validation.Required),
		// validation.Field(&input.EncryptedTransfers, validation.Required),
		// validation.Field(&input.EncryptedDeposits, validation.Required),
		validation.Field(&input.Signature, validation.Required),
	)
}
