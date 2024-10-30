package get_backup_transfer_by_hash

import "github.com/prodadidb/go-validation"

func (input *UCGetBackupTransferByHashInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Recipient, validation.Required),
		validation.Field(&input.TransferHash, validation.Required),
	)
}
