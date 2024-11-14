package intmax_block_content

import "errors"

// ErrMethodNameInvalidStr error: invalid method name: %s
const ErrMethodNameInvalidStr = "invalid method name: %s"

// ErrDecodeCallDataFail error: failed to decode calldata.
var ErrDecodeCallDataFail = errors.New("failed to decode calldata")

// ErrUnpackCalldataFail error: failed to unpack calldata.
var ErrUnpackCalldataFail = errors.New("failed to unpack calldata")

// ErrCannotDecodeAddress error: cannot decode address.
var ErrCannotDecodeAddress = errors.New("cannot decode address")

// ErrSetTxRootFail error: failed to set tx tree root.
var ErrSetTxRootFail = errors.New("failed to set tx tree root")

// ErrRecoverAccountIDsFromBytesFail error: failed to recover account IDs from bytes.
var ErrRecoverAccountIDsFromBytesFail = errors.New("failed to recover account IDs from bytes")

// ErrUnknownAccountID error: account ID is unknown.
var ErrUnknownAccountID = errors.New("account ID is unknown")
