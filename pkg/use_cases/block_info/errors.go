package block_info

import "errors"

// ErrInvalidPrivateKey error: invalid private key.
var ErrInvalidPrivateKey = errors.New("invalid private key")

// ErrStorageGPOValueFail error: failed to get storage value of the gas price oracle.
var ErrStorageGPOValueFail = errors.New("failed to get storage value of the gas price oracle")
