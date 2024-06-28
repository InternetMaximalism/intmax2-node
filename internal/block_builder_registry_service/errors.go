package block_builder_registry_service

import "errors"

// ErrScrollNetworkChainLinkEvmJSONRPCFail error: failed to get the scroll network chain link evm JSON RPC.
var ErrScrollNetworkChainLinkEvmJSONRPCFail = errors.New(
	"failed to get the scroll network chain link evm JSON RPC",
)

// ErrCreateNewClientOfRPCEthFail error: failed to create new RPC Eth client.
var ErrCreateNewClientOfRPCEthFail = errors.New(
	"failed to create new RPC Eth client",
)

// ErrLoadPrivateKeyFail error: failed to load private key.
var ErrLoadPrivateKeyFail = errors.New("failed to load private key")

// ErrParseStrToIntFail error: failed to parse string to integer.
var ErrParseStrToIntFail = errors.New("failed to parse string to integer")

// ErrCreateTransactorFail error: failed to create transactor.
var ErrCreateTransactorFail = errors.New("failed to create transactor")

// ErrCreateOptionsOfTransactionFail error: failed to create options of transactor.
var ErrCreateOptionsOfTransactionFail = errors.New("failed to create options of transactor")

// ErrNewBlockBuilderRegistryCallerFail error: failed to create new the block builder registry caller.
var ErrNewBlockBuilderRegistryCallerFail = errors.New(
	"failed to create new the block builder registry caller",
)

// ErrNewBlockBuilderRegistryTransactorFail error: failed to create new the block builder registry transactor.
var ErrNewBlockBuilderRegistryTransactorFail = errors.New(
	"failed to create new the block builder registry transactor",
)

// ErrProcessingFuncUpdateBlockBuilderOfBlockBuilderRegistryFail error: failed to processing the func 'updateBlockBuilder' of 'block-builder-registry contract'.
var ErrProcessingFuncUpdateBlockBuilderOfBlockBuilderRegistryFail = errors.New(
	"failed to processing the func 'updateBlockBuilder' of 'block-builder-registry contract'",
)

// ErrProcessingFuncStopOfBlockBuilderRegistryFail error: failed to processing the func 'stop' of 'block-builder-registry contract'.
var ErrProcessingFuncStopOfBlockBuilderRegistryFail = errors.New(
	"failed to processing the func 'stop' of 'block-builder-registry contract'",
)

// ErrProcessingFuncUnStakeOfBlockBuilderRegistryFail error: failed to processing the func 'unstake' of 'block-builder-registry contract'.
var ErrProcessingFuncUnStakeOfBlockBuilderRegistryFail = errors.New(
	"failed to processing the func 'stop' of 'block-builder-registry contract'",
)
