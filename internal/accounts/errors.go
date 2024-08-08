package accounts

import "errors"

// ErrHEXPrivateKeyInvalid error: the HEX private key must be valid.
var ErrHEXPrivateKeyInvalid = errors.New("the HEX private key must be valid")

// ErrDecodePrivateKeyFail error: failed to decode private key.
var ErrDecodePrivateKeyFail = errors.New("failed to decode private key")

// ErrCreatePrivateKeyFail error: failed to create new private key.
var ErrCreatePrivateKeyFail = errors.New("failed to create new private key")

// ErrInputPrivateKeyEmpty error: input private key should not be nil.
var ErrInputPrivateKeyEmpty = errors.New("input private key should not be nil")

// ErrInputPrivateKeyInvalid error: input private key should be less than the order of the scalar field.
var ErrInputPrivateKeyInvalid = errors.New("input private key should be less than the order of the scalar field")

// ErrInputPrivateKeyIsZero error: input private key should not be zero.
var ErrInputPrivateKeyIsZero = errors.New("input private key should not be zero")

// ErrPrivateKeyWithPublicKeyInvalid error: invalid private key with the y coordinate of public key not be even number.
var ErrPrivateKeyWithPublicKeyInvalid = errors.New(
	"invalid private key with the y coordinate of public key not be even number",
)

// ErrValidPublicKeyFail error: failed to valid public key.
var ErrValidPublicKeyFail = errors.New("failed to valid public key")

// ErrAddressInvalid error: invalid address.
var ErrAddressInvalid = errors.New("invalid address")

// ErrDecodeAddressFail error: failed to decode address.
var ErrDecodeAddressFail = errors.New("failed to decode address")

// ErrHEXPublicKeyInvalid error: the HEX public key must be valid.
var ErrHEXPublicKeyInvalid = errors.New("the HEX public key must be valid")

// ErrDecodePublicKeyFail error: failed to decode public key.
var ErrDecodePublicKeyFail = errors.New("failed to decode public key")

// ErrUnmarshalPublicKeyFail error: failed to unmarshal public key.
var ErrUnmarshalPublicKeyFail = errors.New("failed to unmarshal public key")
