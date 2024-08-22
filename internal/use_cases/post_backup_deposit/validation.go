package post_backup_deposit

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCPostBackupDepositInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.DepositHash, validation.Required),
		validation.Field(&input.EncryptedDeposit, validation.Required),
		validation.Field(&input.Recipient, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
	)
}
