package post_backup_transaction

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCPostBackupTransactionInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.TxHash, validation.Required),
		validation.Field(&input.EncryptedTx, validation.Required),
		validation.Field(&input.Sender, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
		validation.Field(&input.Signature, validation.Required),
	)
}
