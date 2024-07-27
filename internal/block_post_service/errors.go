package block_post_service

import "errors"

var ErrTransactionByHashNotFound = errors.New("failed to get transaction by hash")

var ErrTransactionIsStillPending = errors.New("transaction is still pending")
