package get_backup_deposits

import "github.com/prodadidb/go-validation"

func (input *UCGetBackupDepositsInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required),
	)
}
