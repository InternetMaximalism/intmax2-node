package tx_transactions_list

import "errors"

// ErrInputValueEmpty error: input value is empty.
var ErrInputValueEmpty = errors.New("input value is empty")

// ErrMarshalJSONFail error: failed to marshal JSON.
var ErrMarshalJSONFail = errors.New("failed to marshal JSON")

// ErrEmptyUserPrivateKey error: user private key is empty.
var ErrEmptyUserPrivateKey = errors.New("user private key is empty")
