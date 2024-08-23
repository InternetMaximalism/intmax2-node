package get_backup_transfers

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCGetBackupTransfersInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required),
	)
}
