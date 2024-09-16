package get_backup_deposit_by_hash

import "github.com/prodadidb/go-validation"

func (input *UCGetBackupDepositByHashInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Recipient, validation.Required),
		validation.Field(&input.DepositHash, validation.Required),
	)
}
