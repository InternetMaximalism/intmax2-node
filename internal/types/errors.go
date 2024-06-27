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

// ErrBlockContentAccIDForAccIDEmpty error: account ID must be non-zero for ACCOUNT_ID sender type.
var ErrBlockContentAccIDForAccIDEmpty = errors.New("account ID must be non-zero for ACCOUNT_ID sender type")

// ErrBlockContentAccIDForDefAccNotEmpty error: account ID must be zero for default sender.
var ErrBlockContentAccIDForDefAccNotEmpty = errors.New("account ID must be zero for default sender")

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
