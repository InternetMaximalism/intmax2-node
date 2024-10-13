package get_backup_transfers_list

import "errors"

// ErrUCGetBackupTransfersInputEmpty error: ucGetBackupTransfersInput must not be empty.
var ErrUCGetBackupTransfersInputEmpty = errors.New("ucGetBackupTransfersInput must not be empty")

// ErrGetBackupTransfersByRecipientFail error: failed to get backup transfers by recipient.
var ErrGetBackupTransfersByRecipientFail = errors.New(
	"failed to get backup transfers by recipient",
)

// ErrJSONMarshalFail error: failed to marshal json.
var ErrJSONMarshalFail = errors.New("failed to marshal json")
