package types

import "errors"

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("value must be valid")

// ErrBlockContentSenderTypeInvalid error: invalid sender type.
var ErrBlockContentSenderTypeInvalid = errors.New("invalid sender type")

// ErrBlockContentSendersEmpty error: no senders.
var ErrBlockContentSendersEmpty = errors.New("no senders")

// ErrBlockContentManySenders error: too many senders.
var ErrBlockContentManySenders = errors.New("too many senders")

// ErrBlockContentPublicKeyNotSorted error: public keys are not sorted.
var ErrBlockContentPublicKeyNotSorted = errors.New("public keys are not sorted")

// ErrBlockContentPublicKeyInvalid error: invalid public key.
var ErrBlockContentPublicKeyInvalid = errors.New("invalid public key")

// ErrBlockContentAccIDForPubKeyInvalid error: account ID must be zero for PUBLIC_KEY sender type.
var ErrBlockContentAccIDForPubKeyInvalid = errors.New("account ID must be zero for PUBLIC_KEY sender type")

// ErrBlockContentAccIDForAccIDInvalid error: account ID must be greater than one for ACCOUNT_ID sender type.
var ErrBlockContentAccIDForAccIDInvalid = errors.New("account ID must be greater than one for ACCOUNT_ID sender type")

// ErrBlockContentAccIDForAccIDEmpty error: account ID must be non-zero for ACCOUNT_ID sender type.
var ErrBlockContentAccIDForAccIDEmpty = errors.New("account ID must be non-zero for ACCOUNT_ID sender type")

// ErrBlockContentAccIDForDefAccNotEmpty error: account ID must be one for default sender.
var ErrBlockContentAccIDForDefAccNotEmpty = errors.New("account ID must be one for default sender")

// ErrBlockContentAggPubKeyEmpty error: no aggregated public key.
var ErrBlockContentAggPubKeyEmpty = errors.New("no aggregated public key")

// ErrBlockContentAggPubKeyInvalid error: invalid aggregated public key.
var ErrBlockContentAggPubKeyInvalid = errors.New("invalid aggregated public key")

// ErrBlockContentAggSignEmpty error: no aggregated signature.
var ErrBlockContentAggSignEmpty = errors.New("no aggregated signature")

// ErrETHAddressInvalid error: the Ethereum address should be 20 bytes.
var ErrETHAddressInvalid = errors.New("the Ethereum address should be 20 bytes")

// ErrINTMAXAddressInvalid error: the INTMAX address should be 32 bytes.
var ErrINTMAXAddressInvalid = errors.New("the INTMAX address should be 32 bytes")

// ErrNonceTooLarge error: nonce is too large.
var ErrNonceTooLarge = errors.New("nonce is too large")

// ErrPowNonceTooLarge error: PoW nonce is too large.
var ErrPowNonceTooLarge = errors.New("PoW nonce is too large")

var ErrTokenTypeRequired = errors.New("token type is required")

var ErrInvalidETHArgs = errors.New("ETH operation requires additional arguments")

var ErrInvalidERC20Args = errors.New("ERC20 operation requires a token address")

var ErrInvalidERC721Args = errors.New("ERC721 operation requires a token address and token ID")

var ErrInvalidERC1155Args = errors.New("ERC1155 operation requires a token address and token ID")

var ErrInvalidTokenType = errors.New("invalid token type. Use 'eth', 'erc20', 'erc721', or 'erc1155'")

var ErrFailedToGetBalance = errors.New("failed to get balance")

var ErrFetchTokenByTokenAddressAndTokenIDWithDBApp = errors.New("failed to fetch token by tokenAddress and tokenId with DBApp")

var ErrFilterLogsFail = errors.New("failed to filter logs")

var ErrEncounteredWhileIterating = errors.New("encountered while iterating error occurred")
