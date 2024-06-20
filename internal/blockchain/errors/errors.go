package errors

import (
	"errors"
)

var InsufficientFunds bool

const ErrInsufficientStakeAmountStr = "Insufficient stake amount"

// ErrInsufficientFundsStr error: insufficient funds: insufficient funds.
const ErrInsufficientFundsStr = "insufficient funds: insufficient funds"

// ErrWalletInsufficientFundsForNodeStart error: wallet insufficient funds for node start.
var ErrWalletInsufficientFundsForNodeStart = errors.New("wallet insufficient funds for node start")

// ErrInvalidSequenceStr error: invalid sequence: invalid sequence.
const ErrInvalidSequenceStr = "invalid sequence: invalid sequence"

// Err520ScrollWebServerStr error: 520: Web server is returning an unknown error.
const Err520ScrollWebServerStr = "520: Web server is returning an unknown error"

// Err502ScrollWebServerStr error: 502: Bad gateway.
const Err502ScrollWebServerStr = "502: Bad gateway"

// ErrStdinProcessingFail error: stdin processing error occurred.
var ErrStdinProcessingFail = errors.New("stdin processing error occurred")

// ErrWalletAddressNotRecognized error: wallet address not recognized.
var ErrWalletAddressNotRecognized = errors.New("wallet address not recognized")

// ErrRecognizingPrivateKeyFail error: failed to recognized private key.
var ErrRecognizingPrivateKeyFail = errors.New("failed to recognized private key")

const ErrScrollChainIDInvalidStr = "the scroll chain ID must be equal: %s, %s"

// ErrParseChainIDFail error: failed to parse chain ID.
var ErrParseChainIDFail = errors.New("failed to parse chain ID")

// ErrSetupScrollNetworkChainIDFail error: failed to setup scroll network chain ID.
var ErrSetupScrollNetworkChainIDFail = errors.New(
	"failed to setup scroll network chain ID",
)

// ErrCreateNewClientOfRPCEthFail error: failed to create new RPC Eth client.
var ErrCreateNewClientOfRPCEthFail = errors.New(
	"failed to create new RPC Eth client",
)

// ErrCreateEthClientFail error: failed to create Eth client.
var ErrCreateEthClientFail = errors.New("failed to create Eth client")

// ErrGetWalletBalanceFail error: failed to get wallet balance.
var ErrGetWalletBalanceFail = errors.New("failed to get wallet balance")

// ErrGettingWalletBalanceErrorOccurred error: getting wallet balance error occurred.
var ErrGettingWalletBalanceErrorOccurred = errors.New("getting wallet balance error occurred")

// ErrScrollNetworkChainLinkEvmJSONRPCFail error: failed to get the chain-link-evm-json-rpc of scroll network.
var ErrScrollNetworkChainLinkEvmJSONRPCFail = errors.New(
	"failed to get the chain-link-evm-json-rpc of scroll network",
)

// ErrMostBlockNumberFail error: failed to get block number.
var ErrMostBlockNumberFail = errors.New("failed to get most block number")

// ErrReadContractTemplateFile error: failed to read contract template file.
var ErrReadContractTemplateFile = errors.New(
	"failed to read contract template file",
)

// ErrGetAbiFail error: failed to get abi.
var ErrGetAbiFail = errors.New("failed to get abi")

// ErrCallContractFromBlockchainFail error: failed to call contract from blockchain.
var ErrCallContractFromBlockchainFail = errors.New("failed to call contract from blockchain")

// ErrCallContractFail error: failed to call contract.
var ErrCallContractFail = errors.New("failed to call contract")

// ErrCallRollupContractFail error: failed to call rollup contract.
var ErrCallRollupContractFail = errors.New("failed to call rollup contract")

// ErrPendingNonceAtFail error: failed to pending nonceAt.
var ErrPendingNonceAtFail = errors.New("failed to pending nonceAt")

// ErrChainIDFormCtxFail error: failed to get chainID from context.
var ErrChainIDFormCtxFail = errors.New("failed to get chainID from context")

// ErrSuggestGasPriceFail error: failed to suggest gas price.
var ErrSuggestGasPriceFail = errors.New("failed to suggest gas price")

// ErrNewKeyedTransactorWithChainIDFail error: failed to get new keyed transactor with chainID.
var ErrNewKeyedTransactorWithChainIDFail = errors.New("failed to get new keyed transactor with chainID")

// ErrApplyBoundContractTransactorFail error: failed to apply bound contract transactor.
var ErrApplyBoundContractTransactorFail = errors.New("failed to apply bound contract transactor")

// ErrApplyContractTransactorFail error: failed to apply contract transactor.
var ErrApplyContractTransactorFail = errors.New("failed to apply contract transactor")

// ErrApplyTransactOfContractRollupFail error: failed to apply transactor of Rollup contract.
var ErrApplyTransactOfContractRollupFail = errors.New("failed to apply transactor of Rollup contract")
