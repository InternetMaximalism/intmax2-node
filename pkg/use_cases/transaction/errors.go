package transaction

import "errors"

// ErrTransferWorkerReceiverFail error: failed to transfer info to w receiver.
var ErrTransferWorkerReceiverFail = errors.New("failed to transfer info to w receiver")
