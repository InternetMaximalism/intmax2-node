package get_backup_transaction_by_hash

import "github.com/prodadidb/go-validation"

func (input *UCGetBackupTransactionByHashInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Sender, validation.Required),
		validation.Field(&input.TxHash, validation.Required),
	)
}
