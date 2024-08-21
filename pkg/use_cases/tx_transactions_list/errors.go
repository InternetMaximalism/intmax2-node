package tx_transactions_list

import "errors"

// ErrEmptyUserPrivateKey error: user private key is empty.
var ErrEmptyUserPrivateKey = errors.New("user private key is empty")

// ErrMoreThenZeroLimit error: limit must be more than zero.
var ErrMoreThenZeroLimit = errors.New("limit must be more than zero")

// ErrInvalidLimit error: limit must be valid value.
var ErrInvalidLimit = errors.New("limit must be valid value")

// ErrInvalidStartBlockNumber error: start block number must be valid.
var ErrInvalidStartBlockNumber = errors.New("start block number must be valid")
