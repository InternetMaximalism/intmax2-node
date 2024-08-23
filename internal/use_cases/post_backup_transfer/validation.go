package post_backup_transfer

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCPostBackupTransferInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.TransferHash, validation.Required),
		validation.Field(&input.EncryptedTransfer, validation.Required),
		validation.Field(&input.Recipient, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
	)
}
