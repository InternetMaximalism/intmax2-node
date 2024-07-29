package block_signature

import "errors"

var ErrInvalidSignature = errors.New("fail to verify signature")

var ErrInvalidEnoughBalanceProof = errors.New("invalid enough balance proof")

var ErrUnmarshalSignatureFail = errors.New("failed to unmarshal signature")

var ErrTooManySenderPublicKeys = errors.New("too many sender public keys")

var ErrInvalidSendersLength = errors.New("invalid senders length")
