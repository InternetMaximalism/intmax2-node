package transaction

import "errors"

// ErrTransferWorkerReceiverFail error: failed to transfer info to worker receiver.
var ErrTransferWorkerReceiverFail = errors.New("failed to transfer info to worker receiver")
