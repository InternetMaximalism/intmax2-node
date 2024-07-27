package block_post_service

import "errors"

// ErrMethodNameInvalidStr error: invalid method name: %s
const ErrMethodNameInvalidStr = "invalid method name: %s"

// ErrTransactionByHashNotFound error: failed to get transaction by hash.
var ErrTransactionByHashNotFound = errors.New("failed to get transaction by hash")

// ErrTransactionIsStillPending error: transaction is still pending.
var ErrTransactionIsStillPending = errors.New("transaction is still pending")

// ErrUnknownAccountID error: account ID is unknown.
var ErrUnknownAccountID = errors.New("account ID is unknown")

// ErrCannotDecodeAddress error: cannot decode address.
var ErrCannotDecodeAddress = errors.New("cannot decode address")

// ErrNewEthereumClientFail error: failed to create new Ethereum client.
var ErrNewEthereumClientFail = errors.New("failed to create new Ethereum client")

// ErrNewScrollClientFail error: failed to create new Scroll client.
var ErrNewScrollClientFail = errors.New("failed to create new Scroll client")

// ErrInstantiateLiquidityContractFail error: failed to instantiate a Liquidity contract.
var ErrInstantiateLiquidityContractFail = errors.New("failed to instantiate a Liquidity contract")

// ErrInstantiateRollupContractFail error: failed to instantiate a Rollup contract.
var ErrInstantiateRollupContractFail = errors.New("failed to instantiate a Rollup contract")

// ErrFilterLogsFail error: failed to filter logs.
var ErrFilterLogsFail = errors.New("failed to filter logs")

// ErrEncounteredWhileIterating error: encountered while iterating error occurred.
var ErrEncounteredWhileIterating = errors.New("encountered while iterating error occurred")

// ErrDecodeCallDataFail error: failed to decode calldata.
var ErrDecodeCallDataFail = errors.New("failed to decode calldata")

// ErrUnpackCalldataFail error: failed to unpack calldata.
var ErrUnpackCalldataFail = errors.New("failed to unpack calldata")

// ErrSetTxRootFail error: failed to set tx tree root.
var ErrSetTxRootFail = errors.New("failed to set tx tree root")

// ErrRecoverAccountIDsFromBytesFail error: failed to recover account IDs from bytes.
var ErrRecoverAccountIDsFromBytesFail = errors.New("failed to recover account IDs from bytes")
