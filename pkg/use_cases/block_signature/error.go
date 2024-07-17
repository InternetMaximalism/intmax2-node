package block_signature

import "errors"

var ErrInvalidSignature = errors.New("fail to verify signature")

var ErrInvalidEnoughBalanceProof = errors.New("invalid enough balance proof")
