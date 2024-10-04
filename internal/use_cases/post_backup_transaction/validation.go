package post_backup_transaction

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCPostBackupTransactionInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.TxHash, validation.Required),
		validation.Field(&input.EncryptedTx, validation.Required),
		validation.Field(&input.SenderEnoughBalanceProofBody, validation.Required),
		// validation.Field(&input.SenderLastBalanceProofBody, validation.Required),
		// validation.Field(&input.SenderBalanceTransitionProofBody, validation.Required),
		validation.Field(&input.Sender, validation.Required),
		validation.Field(&input.BlockNumber, validation.Required),
		validation.Field(&input.Signature, validation.Required),
	)
}

func (input *UCPostBackupTransactionInput) Set(src *UCPostBackupTransactionInput) *UCPostBackupTransactionInput {
	input.TxHash = src.TxHash
	input.EncryptedTx = src.EncryptedTx
	input.SenderEnoughBalanceProofBody = src.SenderEnoughBalanceProofBody
	// input.SenderLastBalanceProofBody = src.SenderLastBalanceProofBody
	// input.SenderBalanceTransitionProofBody = src.SenderBalanceTransitionProofBody
	input.Sender = src.Sender
	input.BlockNumber = src.BlockNumber
	input.Signature = src.Signature

	return input
}
