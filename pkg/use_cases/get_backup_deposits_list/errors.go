package get_backup_deposits_list

import "errors"

// ErrUCGetBackupTransactionsInputEmpty error: ucGetBackupTransactionsInput must not be empty.
var ErrUCGetBackupTransactionsInputEmpty = errors.New("ucGetBackupTransactionsInput must not be empty")

// ErrGetBackupTransactionsBySenderFail error: failed to get backup transactions by sender.
var ErrGetBackupTransactionsBySenderFail = errors.New(
	"failed to get backup transactions by sender",
)

// ErrJSONMarshalFail error: failed to marshal json.
var ErrJSONMarshalFail = errors.New("failed to marshal json")
