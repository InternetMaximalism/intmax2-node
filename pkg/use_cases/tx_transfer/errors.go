package tx_transfer

import "errors"

// ErrUCTransactionRequestEmpty error: uc-transaction-request must not be empty.
var ErrUCTransactionRequestEmpty = errors.New("uc-transaction-request must not be empty")
