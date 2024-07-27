package block_post_service

import "errors"

var ErrTransactionByHashNotFound = errors.New("failed to get transaction by hash")

var ErrTransactionIsStillPending = errors.New("transaction is still pending")

var ErrUnknownAccountID = errors.New("account ID is unknown")

var ErrCannotDecodeAddress = errors.New("cannot decode address")
