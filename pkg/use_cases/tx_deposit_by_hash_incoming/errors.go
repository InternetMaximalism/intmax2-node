package tx_deposit_by_hash_incoming

import "errors"

// ErrEmptyUserPrivateKey error: user private key is empty.
var ErrEmptyUserPrivateKey = errors.New("user private key is empty")

// ErrEmptyTxHash error: transaction hash is empty.
var ErrEmptyTxHash = errors.New("transaction hash is empty")
