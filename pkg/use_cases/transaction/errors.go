package transaction

import "errors"

// ErrUCInputEmpty error: uc-input must not be empty.
var ErrUCInputEmpty = errors.New("uc-input must not be empty")

// ErrTransferWorkerReceiverFail error: failed to transfer info to worker receiver.
var ErrTransferWorkerReceiverFail = errors.New("failed to transfer info to worker receiver")
