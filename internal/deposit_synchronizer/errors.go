package deposit_synchronizer

import "errors"

// ErrInitFail error: failed to init.
var ErrInitFail = errors.New("failed to init")

// ErrNewClientFail error: failed to create new client.
var ErrNewClientFail = errors.New("failed to create new client")
