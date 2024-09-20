package errors

import (
	"errors"
)

var InsufficientFunds bool

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

// Err502EthereumWevServerStr error: 502 Bad Gateway.
const Err502EthereumWevServerStr = "502 Bad Gateway"

// Err503EthereumWebServerStr error: 503 Service Unavailable.
const Err503EthereumWebServerStr = "503 Service Unavailable"

// ErrBlockBuilderNotFoundStr error: Block builder not found.
const ErrBlockBuilderNotFoundStr = "Block builder not found"

// ErrCantUnStakeBlockBuilderStr error: Cannot unstake within one day of the last block submission.
const ErrCantUnStakeBlockBuilderStr = "Cannot unstake within one day of the last block submission"

// ErrInsufficientStakeAmountStr error: Insufficient stake amount.
const ErrInsufficientStakeAmountStr = "Insufficient stake amount"

// ErrStdinProcessingFail error: stdin processing error occurred.
var ErrStdinProcessingFail = errors.New("stdin processing error occurred")

// ErrWalletAddressNotRecognized error: wallet address not recognized.
var ErrWalletAddressNotRecognized = errors.New("wallet address not recognized")

// ErrRecognizingScrollPrivateKeyFail error: failed to recognized the Scroll private key.
var ErrRecognizingScrollPrivateKeyFail = errors.New("failed to recognized the Scroll private key")

// ErrRecognizingEthereumPrivateKeyFail error: failed to recognized the Ethereum private key.
var ErrRecognizingEthereumPrivateKeyFail = errors.New("failed to recognized the Ethereum private key")

// ErrScrollChainIDInvalidStr error: the scroll chain ID must be equal: %s, %s.
const ErrScrollChainIDInvalidStr = "the scroll chain ID must be equal: %s, %s"

// ErrEthereumChainIDInvalidStr error: the ethereum chain ID must be equal: %s, %s.
const ErrEthereumChainIDInvalidStr = "the ethereum chain ID must be equal: %s, %s"

// ErrParseChainIDFail error: failed to parse chain ID.
var ErrParseChainIDFail = errors.New("failed to parse chain ID")

// ErrSetupScrollNetworkChainIDFail error: failed to setup scroll network chain ID.
var ErrSetupScrollNetworkChainIDFail = errors.New(
	"failed to setup scroll network chain ID",
)

// ErrSetupEthereumNetworkChainIDFail error: failed to setup ethereum network chain ID.
var ErrSetupEthereumNetworkChainIDFail = errors.New(
	"failed to setup ethereum network chain ID",
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

// ErrEthClientDialFail error: failed to dial ETH client.
var ErrEthClientDialFail = errors.New("failed to dial ETH client")

// ErrChainIDWithEthClientFail error: failed to get chain ID with ETH client.
var ErrChainIDWithEthClientFail = errors.New("failed to get chain ID with ETH client")
