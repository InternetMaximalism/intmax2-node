package post_backup_transaction

import "errors"

// ErrUCPostBackupTransactionInputEmpty error: ucPostBackupTransactionInput must not be empty.
var ErrUCPostBackupTransactionInputEmpty = errors.New("ucPostBackupTransactionInput must not be empty")
