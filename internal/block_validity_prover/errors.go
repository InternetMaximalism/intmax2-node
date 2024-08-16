package block_validity_prover

import "errors"

// ErrMethodNameInvalidStr error: invalid method name: %s
const ErrMethodNameInvalidStr = "invalid method name: %s"

// ErrCannotDecodeAddress error: cannot decode address.
var ErrCannotDecodeAddress = errors.New("cannot decode address")

// ErrUnknownAccountID error: account ID is unknown.
var ErrUnknownAccountID = errors.New("account ID is unknown")

// ErrDecodeCallDataFail error: failed to decode calldata.
var ErrDecodeCallDataFail = errors.New("failed to decode calldata")

// ErrRecoverAccountIDsFromBytesFail error: failed to recover account IDs from bytes.
var ErrRecoverAccountIDsFromBytesFail = errors.New("failed to recover account IDs from bytes")

// ErrUnpackCalldataFail error: failed to unpack calldata.
var ErrUnpackCalldataFail = errors.New("failed to unpack calldata")

// ErrSetTxRootFail error: failed to set tx tree root.
var ErrSetTxRootFail = errors.New("failed to set tx tree root")

// ErrRegisterPublicKeyFail error: failed to register public key.
var ErrRegisterPublicKeyFail = errors.New("failed to register public key")

var ErrNewEthereumClientFail = errors.New("failed to create new Ethereum client")

var ErrScrollNetwrokChainLink = errors.New("failed to get Scroll network chain link")

var ErrNewScrollClientFail = errors.New("failed to create new Scroll client")

var ErrInstantiateLiquidityContractFail = errors.New("failed to instantiate a Liquidity contract")

var ErrInstantiateRollupContractFail = errors.New("failed to instantiate a Rollup contract")
