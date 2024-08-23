package get_backup_transactions

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCGetBackupTransactionsInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required),
	)
}
