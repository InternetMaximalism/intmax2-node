package block_synchronizer

import "errors"

// ErrTransactionByHashNotFound error: failed to get transaction by hash.
var ErrTransactionByHashNotFound = errors.New("failed to get transaction by hash")

// ErrTransactionIsStillPending error: transaction is still pending.
var ErrTransactionIsStillPending = errors.New("transaction is still pending")

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

// ErrFetchLatestBlockNumberFail error: failed to fetch latest block number.
var ErrFetchLatestBlockNumberFail = errors.New("failed to fetch latest block number")

// ErrNewBlockPostServiceFail error: failed to create new block post service.
var ErrNewBlockPostServiceFail = errors.New("failed to create new block post service")

// ErrNoAssetsFound error: no assets found.
var ErrNoAssetsFound = errors.New("no assets found")
